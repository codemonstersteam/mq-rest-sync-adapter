package team.codemonsters.mqrestsyncadapter.test.stepdefs;

import io.cucumber.datatable.DataTable;
import io.cucumber.java.en.Given;
import io.cucumber.java.en.When;
import io.cucumber.java.en.Then;
import io.cucumber.java.en.And;
import io.restassured.RestAssured;
import io.restassured.response.Response;
import io.restassured.specification.RequestSpecification;
import io.restassured.config.HttpClientConfig;
import io.restassured.config.RestAssuredConfig;
import io.restassured.http.ContentType;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import team.codemonsters.mqrestsyncadapter.test.utils.TestConfig;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.time.LocalDateTime;
import java.time.ZoneOffset;
import java.time.format.DateTimeFormatter;

import static io.restassured.RestAssured.given;
import static org.assertj.core.api.Assertions.assertThat;

public class UnleashStepDefinitions {

    private static final Logger logger = LoggerFactory.getLogger(UnleashStepDefinitions.class);

    private String baseUrl;
    private String adminApiToken;
    private String clientApiToken;
    private String toggleName;
    private String projectName;
    private Response lastResponse;
    private Map<String, Object> toggleDetails = new HashMap<>();
    private Map<String, Object> projectDetails = new HashMap<>();
    private List<String> projectEnvironments = new ArrayList<>();
    private String sessionToken;

    // ========== SERVER AND TOKEN SETUP ==========

    @Given("Unleash server is running")
    public void unleashServerIsRunning() {
        this.baseUrl = "http://" + TestConfig.UNLEASH_HOST + ":" + TestConfig.UNLEASH_PORT;
        this.clientApiToken = TestConfig.UNLEASH_API_TOKEN;
        TestConfig.setupRestAssuredUnleash();

        // Verify server is accessible
        Response healthResponse = given()
                .when()
                .get("/health")
                .then()
                .statusCode(200)
                .extract()
                .response();

        logger.info("Unleash server health check passed: {}", healthResponse.getBody().asString());

        // Create admin token programmatically if needed
        createClientTokenIfNeeded();
    }

    @Given("I have admin API token {string}")
    public void iHaveAdminApiToken(String token) {
        this.adminApiToken = token;
        logger.info("Using admin API token: {}", token);
    }

    @Given("I have default admin API token")
    public void iHaveDefaultAdminApiToken() {
        if (this.adminApiToken == null) {
            createClientTokenIfNeeded();
        }
        logger.info("Using admin API token: {}",
                adminApiToken != null ? "***" + adminApiToken.substring(Math.max(0, adminApiToken.length() - 4))
                        : "null");
    }

    @Given("I have default client API token")
    public void iHaveDefaultClientApiToken() {
        this.clientApiToken = TestConfig.UNLEASH_API_TOKEN;
        logger.info("Using default client API token: {}", clientApiToken);
    }

    private void createClientTokenIfNeeded() {
        // Try to login with default admin credentials and create an admin token
        logger.info("Attempting to create admin token via login...");

        // First, try to get admin token using basic auth with default credentials
        // Default Unleash admin credentials: admin / unleash4all
        Response loginResponse = given()
                .auth().basic("admin", "unleash4all")
                .header("Content-Type", "application/json")
                .body("{\"username\":\"admin\",\"password\":\"unleash4all\"}")
                .when()
                .post("/auth/simple/login")
                .then()
                .extract()
                .response();

        logger.info("Login response: {} - {}", loginResponse.getStatusCode(), loginResponse.getBody().asString());

        if (loginResponse.getStatusCode() == 200) {
            // Get session token from response
            sessionToken = loginResponse.getCookie("unleash-session");
            logger.info("Session token obtained: {}", sessionToken != null ? "yes" : "no");

            if (sessionToken != null) {
                // Create admin API token with correct payload format
                Map<String, Object> tokenPayload = new HashMap<>();
                tokenPayload.put("tokenName", "test-admin-token");
                tokenPayload.put("type", "client");
                String expiresAt = LocalDateTime.now().plusDays(1).atOffset(ZoneOffset.UTC)
                        .format(DateTimeFormatter.ISO_OFFSET_DATE_TIME);
                logger.info("Current date: {}", expiresAt);
                tokenPayload.put("expiresAt", expiresAt);

                logger.info("Creating admin token with payload: {}", tokenPayload);

                Response tokenResponse = given()
                        .cookie("unleash-session", sessionToken)
                        .header("Content-Type", "application/json")
                        .body(tokenPayload)
                        .when()
                        .post("/api/admin/api-tokens")
                        .then()
                        .extract()
                        .response();

                logger.info("Token creation response: {} - {}", tokenResponse.getStatusCode(),
                        tokenResponse.getBody().asString());

                if (tokenResponse.getStatusCode() == 201) {
                    String newAdminToken = tokenResponse.jsonPath().getString("secret");
                    this.adminApiToken = newAdminToken;
                    logger.info("Created new client token successfully");
                }
            }
        }

        logger.info("Failed to create admin token, will use null");
    }

    // ========== PROJECT MANAGEMENT METHODS ==========

    @Given("Project with name {string} exists")
    public void projectDefaultExists(String name) {
        this.projectName = name;
        lastResponse = given()
                .cookie("unleash-session", sessionToken)
                .header("Content-Type", "application/json")
                .when()
                .get("/api/admin/projects")
                .then()
                .extract()
                .response();

        logger.info("Project list response: {} - {}", lastResponse.getStatusCode(), lastResponse.getBody().asString());
        assertThat(lastResponse.getStatusCode()).isEqualTo(200);
        Map<String, Object> toggleMap = TestConfig.jsonToMap(lastResponse.getBody().asString());
    
    }

    @When("I create a feature toggle {string} in the project with type {string}")
    public void iCreateFeatureToggleInProjectWithDetails(String toggleName, String type) {
        this.toggleName = toggleName;
        // Build toggle payload
        Map<String, Object> togglePayload = new HashMap<>();
        togglePayload.put("name", toggleName);
        togglePayload.put("type", type);
        togglePayload.put("impressionData", false);

        logger.info("Creating toggle {} in project {} with payload: {}", toggleName, projectName, togglePayload);

        lastResponse = given()
                .cookie("unleash-session", sessionToken)
                .header("Content-Type", "application/json")
                .body(togglePayload)
                .when()
                .post("/api/admin/projects/" + projectName + "/features")
                .then()
                .extract()
                .response();

        logger.info("Toggle creation response: {} - {}", lastResponse.getStatusCode(),
                lastResponse.getBody().asString());
        assertThat(lastResponse.getStatusCode()).isEqualTo(201);

        // Parse JSON response to validate toggle creation

        String name = lastResponse.jsonPath().getString("name");
        String project = lastResponse.jsonPath().getString("project");
        String toggleType = lastResponse.jsonPath().getString("type");

        // Validate required fields in response
        assertThat(name).isEqualTo(toggleName);
        assertThat(project).isEqualTo(projectName);
        assertThat(toggleType).isEqualTo(type);
        // Log additional fields for debugging
    }

    @Then("the feature toggle should be created successfully in the project")
    public void theFeatureToggleShouldBeCreatedSuccessfullyInProject() {
        assertThat(lastResponse.getStatusCode())
                .as("Feature toggle creation in project should return 201 Created")
                .isEqualTo(201);
    }

    @When("I enable the feature toggle in environment {string}")
    public void iEnableTheFeatureToggleInEnvironment(String environment) {
        logger.info("Enabling toggle {} in project {} for environment {}", toggleName, projectName, environment);

        lastResponse = given()
                .cookie("unleash-session", sessionToken)
                .contentType(ContentType.JSON)
                .body("")
                .when()
                .post("/api/admin/projects/" + projectName + "/features/" + toggleName + "/environments/" + environment + "/on")
                .then()
                .extract()
                .response();

        logger.info("Toggle enable response: {} - {}", lastResponse.getStatusCode(), lastResponse.getBody().asString());
        assertThat(lastResponse.getStatusCode()).isEqualTo(200);
    }

    @Then("the feature toggle should be enabled in environment {string}")
    public void theFeatureToggleShouldBeEnabledInEnvironment(String environment) {
        Response getResponse = given()
                .cookie("unleash-session", sessionToken)
                .when()
                .get("/api/admin/projects/" + projectName + "/features/" + toggleName)
                .then()
                .statusCode(200)
                .extract()
                .response();

        logger.info("Toggle details after enabling: {}", getResponse.getBody().asString());

        // Check if the toggle is enabled in the specified environment
        boolean enabled = getResponse.jsonPath().getBoolean("environments.find { it.name == '" + environment + "' }.enabled");
        assertThat(enabled)
                .as("Toggle should be enabled in environment: " + environment)
                .isTrue();
    }

    @When("I disable the feature toggle in environment {string}")
    public void iDisableTheFeatureToggleInEnvironment(String environment) {
        logger.info("Disabling toggle {} in project {} for environment {}", toggleName, projectName, environment);

        lastResponse = given()
                .cookie("unleash-session", sessionToken)
                .contentType(ContentType.JSON)
                .body("")
                .when()
                .post("/api/admin/projects/" + projectName + "/features/" + toggleName + "/environments/" + environment + "/off")
                .then()
                .extract()
                .response();

        logger.info("Toggle disable response: {} - {}", lastResponse.getStatusCode(), lastResponse.getBody().asString());
        assertThat(lastResponse.getStatusCode()).isEqualTo(200);
    }

    @Then("the feature toggle should be disabled in environment {string}")
    public void theFeatureToggleShouldBeDisabledInEnvironment(String environment) {
        Response getResponse = given()
                .cookie("unleash-session", sessionToken)
                .when()
                .get("/api/admin/projects/" + projectName + "/features/" + toggleName)
                .then()
                .statusCode(200)
                .extract()
                .response();

        logger.info("Toggle details after disabling: {}", getResponse.getBody().asString());

        // Check if the toggle is disabled in the specified environment
        boolean enabled = getResponse.jsonPath().getBoolean("environments.find { it.name == '" + environment + "' }.enabled");
        assertThat(enabled)
                .as("Toggle should be disabled in environment: " + environment)
                .isFalse();
    }

}