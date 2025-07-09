package team.codemonsters.mqrestsyncadapter.test.utils;

import io.restassured.RestAssured;
import io.restassured.config.HttpClientConfig;
import io.restassured.config.RestAssuredConfig;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.core.type.TypeReference;
import java.util.HashMap;

public class TestConfig {
    
    private static final Logger logger = LoggerFactory.getLogger(TestConfig.class);
    
    // MQ Configuration
    public static final String MQ_HOST = getEnv("MQ_HOST", "localhost");
    public static final int MQ_PORT = Integer.parseInt(getEnv("MQ_PORT", "1414"));
    public static final String MQ_QMGR_NAME = getEnv("MQ_QMGR_NAME", "QM1");
    public static final String MQ_CHANNEL = getEnv("MQ_CHANNEL", "DEV.APP.SVRCONN");
    public static final String MQ_USER = getEnv("MQ_USER", "app");
    public static final String MQ_PASSWORD = getEnv("MQ_PASSWORD", "passw0rd");
    
    // WireMock Configuration
    public static final String WIREMOCK_HOST = getEnv("WIREMOCK_HOST", "localhost");
    public static final int WIREMOCK_PORT = Integer.parseInt(getEnv("WIREMOCK_PORT", "8080"));

    // Unleash Configuration
    public static final String UNLEASH_HOST = getEnv("UNLEASH_HOST", "localhost");
    public static final int UNLEASH_PORT = Integer.parseInt(getEnv("UNLEASH_PORT", "4242"));
    public static final String UNLEASH_API_TOKEN = getEnv("UNLEASH_API_TOKEN", "default:development.unleash-insecure-api-token");
    public static final String UNLEASH_ADMIN_API_TOKEN = getEnv("UNLEASH_ADMIN_API_TOKEN", "development.unleash-insecure-admin-api-token");
    
    public static void setupRestAssuredUnleash() {
        // Настройка базового URL из TestConfig
        RestAssured.baseURI = "http://" + UNLEASH_HOST + ":" + UNLEASH_PORT;
        
        // Настройка конфигурации REST Assured
        RestAssured.config = RestAssuredConfig.config()
            .httpClient(HttpClientConfig.httpClientConfig()
                .setParam("http.connection.timeout", 10000)
                .setParam("http.socket.timeout", 10000));
                
        RestAssured.enableLoggingOfRequestAndResponseIfValidationFails();  
        
        logger.info("REST Assured configuration completed for Unleash at {}:{}", 
                   UNLEASH_HOST, UNLEASH_PORT);
    }
    
    private static String getEnv(String key, String defaultValue) {
        String value = System.getenv(key);
        return value != null ? value : defaultValue;
    }

    public static String getUnleashHost() {
        return UNLEASH_HOST;
    }

    public static int getUnleashPort() {
        return UNLEASH_PORT;
    }
    
    public static String getUnleashApiToken() {
        return UNLEASH_API_TOKEN;
    }
    
    public static String getUnleashAdminApiToken() {
        return UNLEASH_ADMIN_API_TOKEN;
    }
    
    public static String getUnleashBaseUrl() {
        return String.format("http://%s:%s", getUnleashHost(), getUnleashPort());
    }

    public static ObjectMapper objectMapper = new ObjectMapper();

    public static HashMap<String,Object> jsonToMap(String json) {
        try {
            return objectMapper.readValue(json, new TypeReference<HashMap<String,Object>>() {});
        } catch (Exception e) {
            throw new RuntimeException("Failed to parse JSON: " + json, e);
        }
    }
}