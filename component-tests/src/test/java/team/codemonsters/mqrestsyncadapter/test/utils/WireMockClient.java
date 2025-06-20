package team.codemonsters.mqrestsyncadapter.test.utils;

import io.restassured.RestAssured;
import io.restassured.response.Response;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class WireMockClient {
    private static final Logger LOG = LoggerFactory.getLogger(WireMockClient.class);
    
    private final String baseUrl;
    
    public WireMockClient(String host, int port) {
        this.baseUrl = "http://" + host + ":" + port;
    }
    
    public boolean isRunning() {
        try {
            Response response = RestAssured.given()
                .when()
                .get(baseUrl + "/__admin/mappings");
            return response.getStatusCode() == 200;
        } catch (Exception e) {
            LOG.error("Error checking WireMock status", e);
            return false;
        }
    }
    
    public String getRequests() {
        try {
            Response response = RestAssured.given()
                .when()
                .get(baseUrl + "/__admin/requests");
            if (response.getStatusCode() == 200) {
                return response.getBody().asString();
            } else {
                LOG.error("Error getting WireMock requests: {}", response.getStatusCode());
                return null;
            }
        } catch (Exception e) {
            LOG.error("Error getting WireMock requests", e);
            return null;
        }
    }
    
    public void resetRequests() {
        try {
            RestAssured.given()
                .when()
                .post(baseUrl + "/__admin/requests/reset");
        } catch (Exception e) {
            LOG.error("Error resetting WireMock requests", e);
        }
    }
} 