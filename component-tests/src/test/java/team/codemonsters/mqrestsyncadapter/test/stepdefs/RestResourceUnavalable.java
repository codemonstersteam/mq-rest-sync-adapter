package team.codemonsters.mqrestsyncadapter.test.stepdefs;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.time.Duration;
import java.time.OffsetDateTime;
import java.util.Arrays;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.Set;
import java.util.stream.Collectors;

import static org.assertj.core.api.Assertions.assertThat;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;

import io.cucumber.java.en.And;
import io.cucumber.java.en.Given;
import io.cucumber.java.en.Then;
import io.cucumber.java.en.When;
import team.codemonsters.mqrestsyncadapter.test.utils.IbmMqClient;

public class RestResourceUnavalable {
    
    protected final ObjectMapper objectMapper = new ObjectMapper();
    protected ObjectNode validPayload;
    protected JsonNode body;
    protected IbmMqClient mqClient;
    protected Set<Map<String, Object>> receipts = new HashSet<>();
    public Map<String, Object> response;

    public RestResourceUnavalable() {
        this.mqClient = new IbmMqClient();
    }
    
    @Given("MQ server is accessible")
    public void mqServerIsAccessible() throws Exception {
        mqClient.setupConnection();
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
        Map<String, String> properties = new HashMap<>();
        properties.put("X_Method", "client-wallet-balance");
        properties.put("X_TransactionID", "transaction-01");
        properties.put("X_RequestID", "x-request-01");
        properties.put("RequestID", "request-01");
        properties.put("CorrelationID", "correlationId");
        
        mqClient.sendMessage(queueName, validPayload.toString(), properties);
    }

    @Then("I read receipt {string}")
    public void readMessage(String queueName) throws Exception {
        Map<String, Object> json = mqClient.readMessage(queueName, 10000);
        assertThat(json).isNotNull();
        receipts.add(json); 
    }

    @Then("I read the error from response queue {string}")
    public void readResponseMessage(String queueName) throws Exception {
        response = mqClient.readMessage(queueName, 5000);
        assertThat(response).isNotNull();
        assertRespone(response);
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