package team.codemonsters.mqrestsyncadapter.test.stepdefs;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.time.Duration;
import java.time.OffsetDateTime;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;

import javax.jms.Connection;
import javax.jms.Message;
import javax.jms.MessageConsumer;
import javax.jms.MessageProducer;
import javax.jms.Queue;
import javax.jms.Session;
import javax.jms.TextMessage;

import static org.assertj.core.api.Assertions.assertThat;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import com.ibm.mq.jms.MQQueueConnectionFactory;

import io.cucumber.java.en.And;
import io.cucumber.java.en.Given;
import io.cucumber.java.en.Then;
import io.cucumber.java.en.When;
import team.codemonsters.mqrestsyncadapter.test.utils.TestConfig;

public class HappyPathFlowSteps {
    private final ObjectMapper objectMapper = new ObjectMapper();
    private ObjectNode validPayload;
    private JsonNode headers;
    private JsonNode body;
    private Message sentMessage;
    private MQQueueConnectionFactory factory;
    private Connection connection;
    private List<TextMessage> receivedReceipts = new ArrayList<TextMessage>(2);
    private TextMessage receivedResponseMessage;
    

    private void setupMqConnection() throws Exception {
        factory = new MQQueueConnectionFactory();
        factory.setHostName(TestConfig.MQ_HOST); // Используем имя сервиса из docker-compose
        factory.setPort(TestConfig.MQ_PORT);
        factory.setQueueManager(TestConfig.MQ_QMGR_NAME);
        factory.setChannel(TestConfig.MQ_CHANNEL); // Должен совпадать с MQ_CHANNEL в компоуз
        factory.setTransportType(1); // Client transport type
        connection = factory.createConnection(TestConfig.MQ_USER, TestConfig.MQ_PASSWORD);

    }

    @Given("MQ server is accessible")
    public void verifyMqServerIsRunning() throws Exception {
        setupMqConnection();
    }

    @And("Create valid message payload from {string}")
    public void verifyMessageValidPayload(String filePath) throws IOException {
        String expectedJson = Files.readString(Paths.get("src/test/resources/files/" + filePath));
        JsonNode message = objectMapper.readTree(expectedJson);
        JsonNode payload = message.get("payload");
        headers = message.get("headers");
        body = payload.get("body");
        validPayload = payload.deepCopy();
        // Модифицируем временные поля
        if (body.has("dateFrom") && body.has("dateTo") && body.has("operTime")) {
            // Создаем копию expectedPayload для модификации
            String jsonString = createValidBody();
            validPayload.put("body", jsonString);
            System.out.println("Valid Payload: " + validPayload.toString());
        } else {
            // Если модификация не удалась (нет нужных полей), возвращаем ошибку
            throw new AssertionError(
                    "Failed to modify payload: required fields (dateFrom, dateTo, operTime) are missing in "
                            + filePath);
        }
    }

    private String createValidBody() {
        OffsetDateTime dateFrom = OffsetDateTime.now().minus(Duration.ofSeconds(10));
        OffsetDateTime dateTo = OffsetDateTime.now().minus(Duration.ofSeconds(5));
        OffsetDateTime operationTime = OffsetDateTime.now().plus(Duration.ofMillis(60000))
                .plus(Duration.ofSeconds(20));

        // Преобразуем в Instant и получаем timestamp с дробной частью
        double dateFromTs = dateFrom.toInstant().getEpochSecond()
                + dateFrom.toInstant().getNano() / 1_000_000_000.0;
        double dateToTs = dateTo.toInstant().getEpochSecond() + dateTo.toInstant().getNano() / 1_000_000_000.0;
        double operTimeTs = operationTime.toInstant().getEpochSecond()
                + operationTime.toInstant().getNano() / 1_000_000_000.0;

        String jsonString = """
                {
                    "requestType": "operationHistory",
                    "dateFrom": %.9f,
                    "dateTo": %.9f,
                    "operTime": %.9f
                }
                """.formatted(dateFromTs, dateToTs, operTimeTs);
        return jsonString;
    }

    @When("I send the message to IN queue {string}")
    public void sendMessage(String queueName) throws Exception {
        if (connection == null) {
            throw new IllegalStateException(
                    "MQ connection is not established. Make sure 'IBM MQ server is running and accessible' step is executed first.");
        }
        try (Session session = connection.createSession()) {
            Queue queue = session.createQueue(queueName);
            MessageProducer producer = session.createProducer(queue);

            TextMessage message = session.createTextMessage(validPayload.toString());
            message.setStringProperty("X_Method", "client-wallet-balance");
            message.setStringProperty("X_TransactionID", "transaction-01");
            message.setStringProperty("X_RequestID", "x-request-01");
            message.setStringProperty("RequestID", "request-01");
            message.setStringProperty("CorrelationID", "correlationId");
            producer.send(message);
            sentMessage = message;
        }
    }

    @Then("I read the message from RECEIPT queue {string} with flag {string}")
    public void readMessage(String queueName, String flag) throws Exception {
        if (connection == null) {
            throw new IllegalStateException(
                    "MQ connection is not established. Make sure 'IBM MQ server is running and accessible' step is executed first.");
        }

        try (Session session = connection.createSession()) {
            Queue queue = session.createQueue(queueName);
            MessageConsumer consumer = session.createConsumer(queue);
            connection.start();
            TextMessage receivedMessage = (TextMessage) consumer.receive(5000);
            System.out.println("Received Message: " + receivedMessage.getText());
            receivedReceipts.add(receivedMessage);
            Map<String, Object> json = objectMapper.readValue(receivedMessage.getText(), new TypeReference<>() {});
            assertReciept(flag, json);
        }
    }
    
    private void assertReciept(String flag, Map<String, Object> json) throws Exception {
        assertThat(json.get("requestType")).isEqualTo("operationHistory");
        assertThat(json.get("flag")).isEqualTo(flag);
        assertThat(json.get("errorCode")).isNull();
    }

    @Then("I read the message from response queue {string}")
    public void readResponseMessage(String queueName) throws Exception {
        if (connection == null) {
            throw new IllegalStateException(
                    "MQ connection is not established. Make sure 'IBM MQ server is running and accessible' step is executed first.");
        }

        try (Session session = connection.createSession()) {
            Queue queue = session.createQueue(queueName);
            MessageConsumer consumer = session.createConsumer(queue);
            connection.start();
            TextMessage receivedMessage = (TextMessage) consumer.receive(5000);
            System.out.println("Received Message: " + receivedMessage.getText());
            receivedResponseMessage = receivedMessage;
            Map<String, Object> json = objectMapper.readValue(receivedMessage.getText(), new TypeReference<>() {});
            assertRespone(json); 
        }
    }

    private void assertRespone(Map<String, Object> json) throws Exception {
        assertThat(json.get("correlationId")).isEqualTo("correlationId");
        assertThat(json.get("status")).isEqualTo(200);
        assertThat(json.get("message")).isEqualTo("OK");
        assertThat(json.get("method")).isEqualTo("POST");
        assertThat(json.get("service")).isEqualTo("api-gateway");
        assertThat(json.get("operation")).isEqualTo("client-wallet-balance");
        Map<String, Object> responseBody = objectMapper.readValue(json.get("body").toString(), new TypeReference<>() {});
        System.out.println("Body: " + responseBody.toString());
        assertThat(responseBody.get("status")).isEqualTo("success");
        assertThat(responseBody.get("actualTimestamp")).isEqualTo(1607449018829L);
        Map<String, Object> data = (Map<String, Object>) responseBody.get("data");
        assertThat(data.get("createTime")).isEqualTo(1600443296227L);
        assertThat(data.get("operationId")).isEqualTo("01");
        Map<String, Object> balance = (Map<String, Object>) data.get("balance");
        assertThat(balance.get("value")).isEqualTo(999.9);
        assertThat(balance.get("currency")).isEqualTo("rub");
    }

}