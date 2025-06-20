package team.codemonsters.mqrestsyncadapter.test.utils;

import org.apache.http.HttpResponse;
import org.apache.http.client.methods.HttpGet;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.util.EntityUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;

public class WireMockClient {
    private static final Logger LOG = LoggerFactory.getLogger(WireMockClient.class);
    
    private final String baseUrl;
    
    public WireMockClient(String host, int port) {
        this.baseUrl = "http://" + host + ":" + port;
    }
    
    public boolean isRunning() {
        try (CloseableHttpClient httpClient = HttpClients.createDefault()) {
            HttpGet request = new HttpGet(baseUrl + "/__admin/mappings");
            HttpResponse response = httpClient.execute(request);
            return response.getStatusLine().getStatusCode() == 200;
        } catch (IOException e) {
            LOG.error("Error checking WireMock status", e);
            return false;
        }
    }
    
    public String getRequests() {
        try (CloseableHttpClient httpClient = HttpClients.createDefault()) {
            HttpGet request = new HttpGet(baseUrl + "/__admin/requests");
            HttpResponse response = httpClient.execute(request);
            if (response.getStatusLine().getStatusCode() == 200) {
                return EntityUtils.toString(response.getEntity());
            } else {
                LOG.error("Error getting WireMock requests: {}", response.getStatusLine());
                return null;
            }
        } catch (IOException e) {
            LOG.error("Error getting WireMock requests", e);
            return null;
        }
    }
    
    public void resetRequests() {
        try (CloseableHttpClient httpClient = HttpClients.createDefault()) {
            HttpGet request = new HttpGet(baseUrl + "/__admin/requests/reset");
            httpClient.execute(request);
        } catch (IOException e) {
            LOG.error("Error resetting WireMock requests", e);
        }
    }
} 