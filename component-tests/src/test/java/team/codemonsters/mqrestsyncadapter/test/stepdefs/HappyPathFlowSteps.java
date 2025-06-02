package team.codemonsters.mqrestsyncadapter.test.stepdefs;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.time.Duration;
import java.time.OffsetDateTime;
import java.util.Map;
import java.util.Set;
import java.util.HashSet;
import java.util.concurrent.TimeUnit;
import java.util.stream.Collectors;

import javax.jms.Connection;
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
    private JsonNode body;
    private MQQueueConnectionFactory factory;
    private Connection connection;
    private static final int MAX_RETRIES = 5;
    private static final int RETRY_DELAY_MS = 100;
    private Set<Map<String, Object>> receipts = new HashSet<Map<String, Object>>();

    private void setupMqConnection() throws Exception {
        factory = new MQQueueConnectionFactory();
        factory.setHostName(TestConfig.MQ_HOST);
        factory.setPort(TestConfig.MQ_PORT);
        factory.setQueueManager(TestConfig.MQ_QMGR_NAME);
        factory.setChannel(TestConfig.MQ_CHANNEL);
        factory.setTransportType(1);
        connection = factory.createConnection(TestConfig.MQ_USER, TestConfig.MQ_PASSWORD);
        connection.start();
    }

    @Given("MQ server is accessible")
    public void mqServerIsAccessible() throws Exception {
        setupMqConnection();
    }

    @And("Create valid message payload from {string}")
    public void verifyMessageValidPayload(String filePath) throws IOException {
        String expectedJson = Files.readString(Paths.get("src/test/resources/files/" + filePath));
        JsonNode message = objectMapper.readTree(expectedJson);
        JsonNode payload = message.get("payload");
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
        OffsetDateTime dateTo = OffsetDateTime.now().plus(Duration.ofSeconds(10));
        OffsetDateTime operationTime = OffsetDateTime.now().minus(Duration.ofMillis(60000));

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
        Session session = connection.createSession();
        Queue queue = session.createQueue(queueName);
        MessageProducer producer = session.createProducer(queue);

        TextMessage message = session.createTextMessage(validPayload.toString());
        message.setStringProperty("X_Method", "client-wallet-balance");
        message.setStringProperty("X_TransactionID", "transaction-01");
        message.setStringProperty("X_RequestID", "x-request-01");
        message.setStringProperty("RequestID", "request-01");
        message.setStringProperty("CorrelationID", "correlationId");
        producer.send(message);
        producer.close();
        TimeUnit.SECONDS.sleep(2);
    }

    @Then("I read the message from RECEIPT queue {string}")
    public void readMessage(String queueName) throws Exception {
        TextMessage receivedMessage = null;
        int retryCount = 0;
        connection = factory.createConnection(TestConfig.MQ_USER, TestConfig.MQ_PASSWORD);
        connection.start();
        Session session = connection.createSession();
        Queue queue = session.createQueue(queueName);
        MessageConsumer consumer = session.createConsumer(queue);
        while (retryCount < MAX_RETRIES) {

            receivedMessage = (TextMessage) consumer.receive(10000);
            if (receivedMessage != null) {
                String messageText = receivedMessage.getText();
                System.out.println("Attempt " + (retryCount + 1) + " - Received message: " + messageText);
                if (messageText != null) {
                    Map<String, Object> json = objectMapper.readValue(messageText, new TypeReference<>() {
                    });
                    receipts.add(json);
                    consumer.close();
                    session.close();
                    connection.close();
                    break;
                }
            }
            retryCount++;
            if (retryCount < MAX_RETRIES) {
                TimeUnit.MILLISECONDS.sleep(RETRY_DELAY_MS);
            }

        }
        consumer.close();
        session.close();
        connection.close();

        if (receivedMessage == null) {
            throw new IllegalStateException("Failed to receive message after " + MAX_RETRIES + " retries");
        }
    }

    @Then("I read the message from response queue {string}")
    public void readResponseMessage(String queueName) throws Exception {
        if (connection == null) {
            throw new IllegalStateException(
                    "MQ connection is not established. Make sure 'IBM MQ server is running and accessible' step is executed first.");
        }

        TextMessage receivedMessage = null;
        int retryCount = 0;
        connection = factory.createConnection(TestConfig.MQ_USER, TestConfig.MQ_PASSWORD);
        connection.start();
        Session session = connection.createSession();
        Queue queue = session.createQueue(queueName);
        MessageConsumer consumer = session.createConsumer(queue);
        while (retryCount < MAX_RETRIES) {
            receivedMessage = (TextMessage) consumer.receive(5000);
            if (receivedMessage != null) {
                String messageText = receivedMessage.getText();
                System.out.println("Attempt " + (retryCount + 1) + " - Received message: " + messageText);
                if (messageText != null) {
                    break;
                }
            }
            retryCount++;
            if (retryCount < MAX_RETRIES) {
                TimeUnit.MILLISECONDS.sleep(RETRY_DELAY_MS);
            }
        }

        if (receivedMessage == null) {
            throw new IllegalStateException("Failed to receive message after " + MAX_RETRIES + " retries");
        }

        System.out.println("Successfully processed message");
        Map<String, Object> json = objectMapper.readValue(receivedMessage.getText(), new TypeReference<>() {
        });
        assertRespone(json);
        consumer.close();
        session.close();
    }

    private void assertRespone(Map<String, Object> json) throws Exception {
        assertThat(json.get("correlationId")).isEqualTo("correlationId");
        assertThat(json.get("status")).isEqualTo(200);
        assertThat(json.get("message")).isEqualTo("OK");
        assertThat(json.get("method")).isEqualTo("POST");
        assertThat(json.get("service")).isEqualTo("api-gateway");
        assertThat(json.get("operation")).isEqualTo("client-wallet-balance");
        Map<String, Object> responseBody = objectMapper.readValue(json.get("body").toString(), new TypeReference<>() {
        });
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

    @Then("I wait for {int} seconds")
    public void waitForSeconds(int seconds) throws Exception {
        TimeUnit.SECONDS.sleep(seconds);
    }

    @Then("I check that i have {int} receipts with flag {string} and {string}")
    public void checkReceipts(int count, String flag1, String flag2) {
        assertThat(receipts).hasSize(count);
        assertThat(receipts.stream()
                .map(json -> (String) json.get("flag"))
                .collect(Collectors.toSet()))
                .containsExactlyInAnyOrder(flag1, flag2);

        receipts.forEach(json -> {
            assertReciept((String) json.get("flag"), json);

        });
    }

    private void assertReciept(String flag, Map<String, Object> json) {
        assertThat(json.get("requestType")).isEqualTo("operationHistory");
        assertThat(json.get("flag")).isEqualTo(flag);
        assertThat(json.get("errorCode")).isNull();
    }

}