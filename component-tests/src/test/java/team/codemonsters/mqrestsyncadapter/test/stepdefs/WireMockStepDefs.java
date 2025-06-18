package team.codemonsters.mqrestsyncadapter.test.stepdefs;

import java.io.IOException;
import java.util.HashMap;
import java.util.Map;

import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.util.EntityUtils;
import static org.assertj.core.api.Assertions.assertThat;
import org.json.JSONObject;
import static org.junit.jupiter.api.Assertions.fail;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.cucumber.java.Before;
import io.cucumber.java.en.Given;
import io.cucumber.java.en.Then;
import io.cucumber.java.en.When;
import team.codemonsters.mqrestsyncadapter.test.client.WireMockClient;
import team.codemonsters.mqrestsyncadapter.test.model.HttpResponse;
import team.codemonsters.mqrestsyncadapter.test.utils.TestConfig;

public class WireMockStepDefs {
    private static final Logger LOG = LoggerFactory.getLogger(WireMockStepDefs.class);
    
    private WireMockClient wireMockClient;
    private HttpResponse response;
    
    @Before
    public void setUp() {
        wireMockClient = new WireMockClient(
            TestConfig.WIREMOCK_HOST,
            TestConfig.WIREMOCK_PORT
        );
    }

    @Given("the WireMock REST API stub is running")
    public void theWireMockRestApiStubIsRunning() {
        assertThat(wireMockClient.isRunning())
            .as("WireMock should be running and accessible")
            .isTrue();
        
        wireMockClient.resetRequests();
    }

    @When("I send a POST request to {string}")
    public void iSendPostRequestTo(String path) {
        try (CloseableHttpClient httpClient = HttpClients.createDefault()) {
            HttpPost request = new HttpPost("http://" + TestConfig.WIREMOCK_HOST + ":" + TestConfig.WIREMOCK_PORT + path);
            request.setHeader("Content-Type", "application/json");
            request.setEntity(new StringEntity("{}"));
            
            try (CloseableHttpResponse httpResponse = httpClient.execute(request)) {
                Map<String, String> headers = new HashMap<>();
                for (org.apache.http.Header header : httpResponse.getAllHeaders()) {
                    headers.put(header.getName(), header.getValue());
                }
                
                response = new HttpResponse(
                    httpResponse.getStatusLine().getStatusCode(),
                    EntityUtils.toString(httpResponse.getEntity()),
                    headers
                );
            }
        } catch (IOException e) {
            LOG.error("Failed to send POST request", e);
            fail("Failed to send POST request: " + e.getMessage());
        }
    }

    @Then("I should receive a successful response")
    public void iShouldReceiveSuccessfulResponse() {
        assertThat(response).isNotNull();
        assertThat(response.getStatusCode()).isEqualTo(200);
    }

    @Then("the response should contain status {string}")
    public void theResponseShouldContainStatus(String status) {
        assertThat(response.getJsonResponse().getString("status")).isEqualTo(status);
    }

    @Then("the response should contain actualTimestamp {string}")
    public void theResponseShouldContainActualTimestamp(String timestamp) {
        assertThat(response.getJsonResponse().getLong("actualTimestamp")).isEqualTo(Long.parseLong(timestamp));
    }

    @Then("the response should contain data with createTime {string}")
    public void theResponseShouldContainDataWithCreateTime(String createTime) {
        assertThat(response.getDataObject().getLong("createTime")).isEqualTo(Long.parseLong(createTime));
    }

    @Then("the response should contain data with operationId {string}")
    public void theResponseShouldContainDataWithOperationId(String operationId) {
        assertThat(response.getDataObject().getString("operationId")).isEqualTo(operationId);
    }

    @Then("the response should contain balance with value {string} and currency {string}")
    public void theResponseShouldContainBalanceWithValueAndCurrency(String value, String currency) {
        JSONObject balance = response.getDataObject().getJSONObject("balance");
        assertThat(balance.getDouble("value")).isEqualTo(Double.parseDouble(value));
        assertThat(balance.getString("currency")).isEqualTo(currency);
    }

    @Then("the response should contain Content-Type header {string}")
    public void theResponseShouldContainContentTypeHeader(String contentType) {
        assertThat(response.getHeaders().get("Content-Type")).isEqualTo(contentType);
    }
} 