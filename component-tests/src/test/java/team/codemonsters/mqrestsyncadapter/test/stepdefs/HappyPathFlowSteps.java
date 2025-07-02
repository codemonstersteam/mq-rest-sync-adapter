package team.codemonsters.mqrestsyncadapter.test.stepdefs;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.time.Duration;
import java.time.OffsetDateTime;
import java.util.HashMap;
import java.util.HashSet;
import java.util.Map;
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

public class HappyPathFlowSteps {
    protected final ObjectMapper objectMapper = new ObjectMapper();
    protected ObjectNode validPayload;
    protected JsonNode body;
    protected IbmMqClient mqClient;
    protected Set<Map<String, Object>> receipts = new HashSet<>();
    public Map<String, Object> response;

    public HappyPathFlowSteps() {
        this.mqClient = new IbmMqClient();
    }

    @Given("MQ server is online")
    public void mqServerIsAccessible() throws Exception {
        mqClient.setupConnection();
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

    @When("I send the message to {string}")
    public void sendMessage(String queueName) throws Exception {
        Map<String, String> properties = new HashMap<>();
        properties.put("X_Method", "client-wallet-balance");
        properties.put("X_TransactionID", "transaction-01");
        properties.put("X_RequestID", "x-request-01");
        properties.put("RequestID", "request-01");
        properties.put("CorrelationID", "correlationId");
        
        mqClient.sendMessage(queueName, validPayload.toString(), properties);
    }

    @Then("I read the message from RECEIPT queue {string}")
    public void readMessage(String queueName) throws Exception {
        Map<String, Object> json = mqClient.readMessage(queueName, 10000);
        assertThat(json).isNotNull();
        receipts.add(json);
    }

    @Then("I read the message from response queue {string}")
    public void readResponseMessage(String queueName) throws Exception {
        response = mqClient.readMessage(queueName, 5000);
        assertThat(response).isNotNull();
        assertRespone();
    }

    private void assertRespone() throws Exception {
        assertThat(response.get("correlationId")).isEqualTo("correlationId");
        assertThat(response.get("status")).isEqualTo(200);
        assertThat(response.get("message")).isEqualTo("OK");
        assertThat(response.get("method")).isEqualTo("POST");
        assertThat(response.get("service")).isEqualTo("api-gateway");
        assertThat(response.get("operation")).isEqualTo("client-wallet-balance");
        Map<String, Object> responseBody = objectMapper.readValue(response.get("body").toString(), new TypeReference<>() {
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

    @Then("I check that i have 2 receipts with flag SUCCESS and ACCEPTED")
    public void checkReceipts() {
        int count = 2;
        String flag1 = "ACCEPTED";
        String flag2 = "SUCCESS";
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

    @Then("I clear receipts state")
    public void clearReceiptsState() {
        receipts.clear();
    }
}