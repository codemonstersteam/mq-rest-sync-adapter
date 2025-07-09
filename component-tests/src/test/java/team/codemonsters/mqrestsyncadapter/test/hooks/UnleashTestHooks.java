package team.codemonsters.mqrestsyncadapter.test.hooks;


import io.cucumber.java.Before;
import io.cucumber.java.After;
import io.cucumber.java.Scenario;
import team.codemonsters.mqrestsyncadapter.test.utils.TestConfig;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import static io.restassured.RestAssured.given;

public class UnleashTestHooks {
    
    private static final Logger logger = LoggerFactory.getLogger(UnleashTestHooks.class);
    
    @Before("@test-unleash")
    public void setupUnleashTest(Scenario scenario) {
        logger.info("Setting up Unleash test: {}", scenario.getName());
        TestConfig.setupRestAssuredUnleash();
        
        // Verify Unleash is accessible
        String baseUrl = TestConfig.getUnleashBaseUrl();
        try {
            given()
                .when()
                .get(baseUrl + "/health")
                .then()
                .statusCode(200);
            logger.info("Unleash server is accessible at: {}", baseUrl);
        } catch (Exception e) {
            logger.error("Failed to connect to Unleash server at: {}", baseUrl, e);
            throw new RuntimeException("Unleash server is not accessible", e);
        }
    }
    
    @After("@test-unleash")
    public void teardownUnleashTest(Scenario scenario) {
        logger.info("Tearing down Unleash test: {} - Status: {}", 
                   scenario.getName(), 
                   scenario.getStatus());
        
        if (scenario.isFailed()) {
            logger.error("Test failed: {}", scenario.getName());
            // Add any additional cleanup for failed tests
        }
    }
    
    @After("@cleanup")
    public void cleanupFeatureToggles(Scenario scenario) {
        logger.info("Cleaning up feature toggles after test: {}", scenario.getName());
        
        String adminToken = TestConfig.getUnleashApiToken();
        String baseUrl = TestConfig.getUnleashBaseUrl();
        
        // Clean up any test toggles that might have been left behind
        try {
            given()
                .header("Authorization", adminToken)
                .when()
                .delete(baseUrl + "/api/admin/projects/default/features/mq-rest-sync-adapter.wallet-balance-workflow")
                .then()
                .statusCode(200);
            logger.info("Cleaned up test toggle: mq-rest-sync-adapter.wallet-balance-workflow");
        } catch (Exception e) {
            logger.warn("Failed to cleanup test toggle (might not exist): {}", e.getMessage());
        }
    }
}