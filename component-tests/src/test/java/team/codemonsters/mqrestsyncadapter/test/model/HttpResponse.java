package team.codemonsters.mqrestsyncadapter.test.model;

import java.util.Collections;
import java.util.Map;

import org.json.JSONObject;

public class HttpResponse {
    private final int statusCode;
    private final String body;
    private final Map<String, String> headers;
    private final JSONObject jsonResponse;

    public HttpResponse(int statusCode, String body, Map<String, String> headers) {
        this.statusCode = statusCode;
        this.body = body;
        this.headers = Collections.unmodifiableMap(headers);
        this.jsonResponse = new JSONObject(body);
    }

    public int getStatusCode() {
        return statusCode;
    }

    public String getBody() {
        return body;
    }

    public Map<String, String> getHeaders() {
        return headers;
    }

    public JSONObject getJsonResponse() {
        return jsonResponse;
    }

    public JSONObject getDataObject() {
        return jsonResponse.getJSONObject("data");
    }
} 