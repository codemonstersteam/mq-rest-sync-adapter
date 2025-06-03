package team.codemonsters.mqrestsyncadapter.test.stepdefs;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.time.Duration;
import java.time.OffsetDateTime;
import java.util.Arrays;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.Set;
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

public class RestResourceUnavalable {
    
    protected final ObjectMapper objectMapper = new ObjectMapper();
    protected ObjectNode validPayload;
    protected JsonNode body;
    protected MQQueueConnectionFactory factory;
    protected Connection connection;
    protected Set<Map<String, Object>> receipts = new HashSet<>();
    public Map<String, Object> response;
    public static final int MAX_RETRIES = 5;
    public static final int RETRY_DELAY_MS = 100;

    public  void setupMqConnection() throws Exception {
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

    @And("Create valid message payload from {string} for test REST resource unavalable")
    public void verifyMessageValidPayload(String filePath) throws IOException {
        String expectedJson = Files.readString(Paths.get("src/test/resources/files/" + filePath));
        JsonNode message = objectMapper.readTree(expectedJson);
        JsonNode payload = message.get("payload");
        body = payload.get("body");
        validPayload = payload.deepCopy();
        JsonNode headers = payload.get("headers");
        ((ObjectNode) headers).put("Api-clientId", "Api-clientId-02");
        ((ObjectNode) headers).put("RequestID", "request-01");
        // Модифицируем временные поля
        if (body.has("dateFrom") && body.has("dateTo") && body.has("operTime")) {
            // Создаем копию expectedPayload для модификации
            String jsonString = createValidBody();
            validPayload.put("body", jsonString);
            validPayload.put("headers", headers);
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

    @When("I send the message with header Api-clientId = Api-clientId-02 to in queue {string}")
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

    @Then("I read receipt {string}")
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

    @Then("I read the error from response queue {string}")
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
        assertThat(json)
            .containsEntry("correlationId", "correlationId")
            .containsEntry("status", 500)
            .containsEntry("message", "Internal Server Error")
            .containsEntry("method", "POST")
            .containsEntry("service", "api-gateway")
            .containsEntry("operation", "client-wallet-balance");

        Map<String, Object> body = objectMapper.readValue(json.get("body").toString(), new TypeReference<>() {});
        assertThat(body)
            .containsEntry("status", "error")
            .hasEntrySatisfying("error", error -> {
                Map<String, Object> errorMap = (Map<String, Object>) error;
                assertThat(errorMap)
                    .containsEntry("errorCode", "500")
                    .containsEntry("errorMessage", "Internal Server Error");
            });

        Map<String, Object> headers = (Map<String, Object>) json.get("headers");
        assertThat(headers)
            .containsEntry("Content-Type", Arrays.asList("application/json"))
            .containsEntry("Transfer-Encoding", Arrays.asList("chunked"))
            .hasEntrySatisfying("Matched-Stub-Id", stubId -> 
                assertThat(stubId).isInstanceOf(List.class).asList().hasSize(1));
    }

    @Then("I check that i have 2 receipts with flag ACCEPTED and ERROR")
    public void checkReceipts() throws Exception {
        int count = 2;
        String acceptedFlag = "ACCEPTED";
        String errorFlag = "ERROR";
        assertThat(receipts).hasSize(count);
        assertThat(receipts.stream()
                .map(json -> (String) json.get("flag"))
                .collect(Collectors.toSet()))
                .containsExactlyInAnyOrder(acceptedFlag, errorFlag);
    
        assertAcceptedReciept(findReciept(acceptedFlag).get());
        assertErrorReciept(findReciept(errorFlag).get());
    }

    private Optional<Map<String, Object>> findReciept(String flag) {
        return receipts.stream().filter(json -> json.get("flag").equals(flag)).findFirst();
    }

    private void assertAcceptedReciept(Map<String, Object> json) {
        assertThat(json.get("requestType")).isEqualTo("operationHistory");
        assertThat(json.get("flag")).isEqualTo("ACCEPTED");
        assertThat(json.get("errorCode")).isNull();
    }

    private void assertErrorReciept(Map<String, Object> json) throws Exception {
        assertThat(json)
            .containsEntry("requestType", "operationHistory")
            .containsEntry("flag", "ERROR")
            .containsEntry("errorCode", 500);
        
        Map<String, Object> description = objectMapper.readValue(json.get("description").toString(), new TypeReference<>() {});
        assertThat(description)
            .containsEntry("status", "error")
            .hasEntrySatisfying("error", error -> {
                Map<String, Object> errorMap = (Map<String, Object>) error;
                assertThat(errorMap)
                    .containsEntry("errorCode", "500")
                    .containsEntry("errorMessage", "Internal Server Error");
            });
    }

}