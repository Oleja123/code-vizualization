package flowchart.parser;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import flowchart.ast.Program;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;

/**
 * Клиент для semantic-analyzer-service API
 * Парсит C код в AST с валидацией
 */
public class SemanticAnalyzerClient {

    private final String apiUrl;
    private final HttpClient httpClient;
    private final ObjectMapper objectMapper;

    public SemanticAnalyzerClient(String apiUrl) {
        this.apiUrl = apiUrl;
        this.httpClient = HttpClient.newBuilder()
                .connectTimeout(Duration.ofSeconds(10))
                .build();
        this.objectMapper = new ObjectMapper();
    }

    /**
     * Парсит C код в AST
     * @param code C код
     * @return Program AST
     * @throws ParseException если парсинг не удался
     */
    public Program parse(String code) throws ParseException {
        try {
            // Формируем JSON запрос
            String requestBody = objectMapper.writeValueAsString(
                    new ValidateRequest(code)
            );

            // Отправляем POST /validate
            HttpRequest request = HttpRequest.newBuilder()
                    .uri(URI.create(apiUrl + "/validate"))
                    .header("Content-Type", "application/json")
                    .timeout(Duration.ofSeconds(30))
                    .POST(HttpRequest.BodyPublishers.ofString(requestBody))
                    .build();

            HttpResponse<String> response = httpClient.send(
                    request,
                    HttpResponse.BodyHandlers.ofString()
            );

            // Логируем для отладки
            System.out.println("[SemanticAnalyzerClient] Status: " + response.statusCode());
            System.out.println("[SemanticAnalyzerClient] Body: " + response.body());

            // Парсим ответ
            JsonNode responseJson = objectMapper.readTree(response.body());

            // Проверяем success
            boolean success = responseJson.path("success").asBoolean(false);

            if (!success) {
                String error = responseJson.path("error").asText("Unknown error");
                throw new ParseException(error);
            }

            // Извлекаем program
            JsonNode programNode = responseJson.path("program");
            if (programNode.isMissingNode()) {
                throw new ParseException("Response missing 'program' field");
            }

            System.out.println("[SemanticAnalyzerClient] Program node type: " + programNode.path("type").asText());

            // Десериализуем Program
            return objectMapper.treeToValue(programNode, Program.class);

        } catch (ParseException e) {
            throw e;
        } catch (Exception e) {
            throw new ParseException("Failed to parse code: " + e.getMessage(), e);
        }
    }

    /**
     * Проверяет здоровье сервиса
     * @return true если сервис доступен
     */
    public boolean isHealthy() {
        try {
            HttpRequest request = HttpRequest.newBuilder()
                    .uri(URI.create(apiUrl + "/health"))
                    .timeout(Duration.ofSeconds(5))
                    .GET()
                    .build();

            HttpResponse<String> response = httpClient.send(
                    request,
                    HttpResponse.BodyHandlers.ofString()
            );

            return response.statusCode() == 200;
        } catch (Exception e) {
            return false;
        }
    }

    /**
     * Получает информацию о сервисе
     * @return JSON с информацией
     */
    public JsonNode getInfo() throws Exception {
        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(apiUrl + "/info"))
                .timeout(Duration.ofSeconds(5))
                .GET()
                .build();

        HttpResponse<String> response = httpClient.send(
                request,
                HttpResponse.BodyHandlers.ofString()
        );

        return objectMapper.readTree(response.body());
    }

    // DTO для запроса
    private static class ValidateRequest {
        public String code;

        public ValidateRequest(String code) {
            this.code = code;
        }
    }
}