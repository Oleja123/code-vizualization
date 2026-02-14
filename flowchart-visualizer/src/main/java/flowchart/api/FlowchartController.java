package flowchart.api;

import flowchart.FlowchartGenerator;
import flowchart.parser.SemanticAnalyzerClient;
import flowchart.parser.ParseException;
import flowchart.ast.Program;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.Map;

/**
 * REST API контроллер для генерации блок-схем
 *
 * Endpoints:
 * POST /api/flowchart/generate - Генерация SVG из AST
 * POST /api/flowchart/generate-from-code - Генерация SVG из C кода (с парсингом через semantic-analyzer)
 */
@RestController
@RequestMapping("/api/flowchart")
@CrossOrigin(origins = "*")
public class FlowchartController {

    private final FlowchartGenerator generator;
    private final SemanticAnalyzerClient astClient;

    public FlowchartController(
            @Value("${ast.service.url:http://localhost:8080}") String astServiceUrl) {
        this.generator = new FlowchartGenerator();
        this.astClient = new SemanticAnalyzerClient(astServiceUrl);
    }

    /**
     * GET /api/flowchart
     * Информация об API
     */
    @GetMapping(produces = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<Map<String, Object>> apiInfo() {
        Map<String, Object> info = new HashMap<>();
        info.put("service", "Flowchart Visualizer");
        info.put("version", "2.0.0");
        info.put("status", "running");
        info.put("ast_service_healthy", astClient.isHealthy());
        return ResponseEntity.ok(info);
    }

    /**
     * POST /api/flowchart/generate
     *
     * Request Body:
     * {
     *   "ast": { ... AST JSON от Go сервиса ... }
     * }
     *
     * Response:
     * {
     *   "svg": "<svg>...</svg>",
     *   "metadata": {
     *     "nodeCount": 10,
     *     "success": true
     *   }
     * }
     */
    @PostMapping(value = "/generate",
            consumes = MediaType.APPLICATION_JSON_VALUE,
            produces = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<Map<String, Object>> generateFlowchart(@RequestBody Map<String, Object> request) {
        try {
            // Извлекаем AST из запроса
            Object astObject = request.get("ast");
            if (astObject == null) {
                return ResponseEntity.badRequest()
                        .body(createErrorResponse("Missing 'ast' field in request"));
            }

            // Конвертируем AST обратно в JSON строку
            com.fasterxml.jackson.databind.ObjectMapper mapper = new com.fasterxml.jackson.databind.ObjectMapper();
            String astJson = mapper.writeValueAsString(astObject);

            // Генерируем SVG
            String svg = generator.generateSVG(astJson);

            // Формируем ответ
            Map<String, Object> response = new HashMap<>();
            response.put("svg", svg);

            Map<String, Object> metadata = new HashMap<>();
            metadata.put("success", true);
            metadata.put("svgLength", svg.length());
            response.put("metadata", metadata);

            return ResponseEntity.ok(response);

        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body(createErrorResponse("Error generating flowchart: " + e.getMessage()));
        }
    }

    /**
     * POST /api/flowchart/generate-from-code
     *
     * Request Body:
     * {
     *   "code": "int main() { return 0; }"
     * }
     *
     * Response:
     * {
     *   "svg": "<svg>...</svg>",
     *   "ast": { ... },
     *   "metadata": {
     *     "success": true,
     *     "nodeCount": 10
     *   }
     * }
     *
     * Интеграция с semantic-analyzer-service:
     * 1. Отправляет код на парсинг в Go сервис
     * 2. Получает валидированный AST
     * 3. Генерирует SVG блок-схему
     */
    @PostMapping(value = "/generate-from-code",
            consumes = MediaType.APPLICATION_JSON_VALUE,
            produces = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<Map<String, Object>> generateFromCode(@RequestBody Map<String, String> request) {
        try {
            String code = request.get("code");
            if (code == null || code.trim().isEmpty()) {
                return ResponseEntity.badRequest()
                        .body(createErrorResponse("Missing 'code' field in request"));
            }

            // 1. Парсим код через semantic-analyzer API
            Program program;
            try {
                program = astClient.parse(code);
            } catch (ParseException e) {
                return ResponseEntity.badRequest()
                        .body(createErrorResponse("Parse error: " + e.getMessage()));
            }

            // 2. Конвертируем Program в JSON
            com.fasterxml.jackson.databind.ObjectMapper mapper = new com.fasterxml.jackson.databind.ObjectMapper();
            String astJson = mapper.writeValueAsString(program);

            // 3. Генерируем SVG
            String svg = generator.generateSVG(astJson);

            // 4. Формируем ответ
            Map<String, Object> response = new HashMap<>();
            response.put("svg", svg);
            response.put("ast", program);

            Map<String, Object> metadata = new HashMap<>();
            metadata.put("success", true);
            metadata.put("svgLength", svg.length());
            response.put("metadata", metadata);

            return ResponseEntity.ok(response);

        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body(createErrorResponse("Error generating flowchart from code: " + e.getMessage()));
        }
    }

    /**
     * GET /api/flowchart/health
     * Проверка работоспособности сервиса
     */
    @GetMapping("/health")
    public ResponseEntity<Map<String, Object>> health() {
        Map<String, Object> response = new HashMap<>();
        response.put("status", "ok");
        response.put("service", "Flowchart Visualizer");
        response.put("version", "2.0.0");
        response.put("ast_service_healthy", astClient.isHealthy());
        return ResponseEntity.ok(response);
    }

    /**
     * Создать ответ с ошибкой
     */
    private Map<String, Object> createErrorResponse(String message) {
        Map<String, Object> response = new HashMap<>();
        Map<String, Object> metadata = new HashMap<>();
        metadata.put("success", false);
        metadata.put("error", message);
        response.put("metadata", metadata);
        return response;
    }
}