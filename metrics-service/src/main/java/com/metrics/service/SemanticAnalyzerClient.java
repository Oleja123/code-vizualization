package com.metrics.service;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.metrics.ast.Program;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;

@Slf4j
@Component
public class SemanticAnalyzerClient {

    private final String baseUrl;
    private final HttpClient httpClient;
    private final ObjectMapper objectMapper;

    public SemanticAnalyzerClient(
            @Value("${ast.service.url:http://localhost:8082}") String baseUrl,
            ObjectMapper objectMapper) {
        this.baseUrl = baseUrl;
        this.objectMapper = objectMapper;
        this.httpClient = HttpClient.newBuilder()
                .connectTimeout(Duration.ofSeconds(10))
                .build();
    }

    public Program parse(String code) throws Exception {
        String body = objectMapper.writeValueAsString(java.util.Map.of("code", code));

        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(baseUrl + "/validate"))
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(body))
                .timeout(Duration.ofSeconds(30))
                .build();

        HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());

        if (response.statusCode() != 200) {
            JsonNode error = objectMapper.readTree(response.body());
            String msg = error.path("error").asText("Parse error");
            throw new RuntimeException(msg);
        }

        JsonNode root = objectMapper.readTree(response.body());
        JsonNode programNode = root.path("program");
        return objectMapper.treeToValue(programNode, Program.class);
    }
}
