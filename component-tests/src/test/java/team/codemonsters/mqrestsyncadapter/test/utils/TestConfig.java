package team.codemonsters.mqrestsyncadapter.test.utils;

public class TestConfig {
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
    
    private static String getEnv(String key, String defaultValue) {
        String value = System.getenv(key);
        return value != null ? value : defaultValue;
    }
} 