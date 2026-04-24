package flowchart.integration;

import flowchart.FlowchartGenerator;
import flowchart.builder.FlowchartBuilder;
import flowchart.renderer.SVGRenderer;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;

import static org.hamcrest.Matchers.*;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.*;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

/**
 * Интеграционные тесты flowchart-visualizer.
 *
 * SemanticAnalyzerClient мокается (внешний Go-сервис недоступен в тестах).
 * FlowchartGenerator, FlowchartBuilder, SVGRenderer работают по-настоящему.
 * Spring Security разрешает все запросы (permitAll в SecurityConfig).
 *
 * Зависимости не нужны дополнительные — spring-boot-starter-test уже есть в pom.xml.
 */
@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
@AutoConfigureMockMvc
@DisplayName("FlowchartController — интеграционные тесты")
class FlowchartControllerIntegrationTest {

    @Autowired
    private MockMvc mockMvc;

    // ─────────────────────────────────────────────────────────────
    //  GET /api/flowchart/health
    // ─────────────────────────────────────────────────────────────

    @Test
    @DisplayName("GET /health → 200 OK, service=Flowchart Visualizer")
    void health() throws Exception {
        mockMvc.perform(get("/api/flowchart/health"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.status").value("ok"))
                .andExpect(jsonPath("$.service").value("Flowchart Visualizer"));
    }

    // ─────────────────────────────────────────────────────────────
    //  GET /api/flowchart (info)
    // ─────────────────────────────────────────────────────────────

    @Test
    @DisplayName("GET /api/flowchart → 200, version=2.0.0")
    void apiInfo() throws Exception {
        mockMvc.perform(get("/api/flowchart"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.service").value("Flowchart Visualizer"))
                .andExpect(jsonPath("$.version").value("2.0.0"));
    }

    // ─────────────────────────────────────────────────────────────
    //  POST /api/flowchart/generate (из готового AST)
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("POST /api/flowchart/generate")
    class Generate {

        @Test
        @DisplayName("Корректный AST → 200, содержит svg и metadata.success=true")
        void validAst() throws Exception {
            String body = """
                    {
                      "ast": {
                        "type": "Program",
                        "declarations": [{
                          "type": "FunctionDecl",
                          "name": "main",
                          "returnType": {"baseType":"int","pointerLevel":0,"arraySizes":[]},
                          "parameters": [],
                          "body": {
                            "type": "BlockStmt",
                            "statements": [{
                              "type": "ReturnStmt",
                              "value": {"type":"IntLiteral","value":0,
                                        "location":{"line":2,"column":5,"endLine":2,"endColumn":6}},
                              "location": {"line":2,"column":5,"endLine":2,"endColumn":6}
                            }],
                            "location": {"line":1,"column":14,"endLine":3,"endColumn":1}
                          },
                          "location": {"line":1,"column":1,"endLine":3,"endColumn":1}
                        }],
                        "location": {"line":1,"column":1,"endLine":3,"endColumn":1}
                      }
                    }
                    """;

            mockMvc.perform(post("/api/flowchart/generate")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content(body))
                    .andExpect(status().isOk())
                    .andExpect(jsonPath("$.svg").exists())
                    .andExpect(jsonPath("$.svg", containsString("<svg")))
                    .andExpect(jsonPath("$.svg", containsString("main")))
                    .andExpect(jsonPath("$.metadata.success").value(true))
                    .andExpect(jsonPath("$.metadata.svgLength").isNumber());
        }

        @Test
        @DisplayName("Отсутствует поле ast → 400")
        void missingAstField() throws Exception {
            mockMvc.perform(post("/api/flowchart/generate")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content("{\"wrong\":\"field\"}"))
                    .andExpect(status().isBadRequest())
                    .andExpect(jsonPath("$.metadata.success").value(false));
        }

        @Test
        @DisplayName("SVG содержит имя функции из AST")
        void svgContainsFunctionName() throws Exception {
            String body = """
                    {
                      "ast": {
                        "type": "Program",
                        "declarations": [{
                          "type": "FunctionDecl",
                          "name": "factorial",
                          "returnType": {"baseType":"int","pointerLevel":0,"arraySizes":[]},
                          "parameters": [{
                            "type": {"baseType":"int","pointerLevel":0,"arraySizes":[]},
                            "name": "n",
                            "location": {"line":1,"column":20,"endLine":1,"endColumn":25}
                          }],
                          "body": {
                            "type": "BlockStmt",
                            "statements": [],
                            "location": {"line":1,"column":28,"endLine":3,"endColumn":1}
                          },
                          "location": {"line":1,"column":1,"endLine":3,"endColumn":1}
                        }],
                        "location": {"line":1,"column":1,"endLine":3,"endColumn":1}
                      }
                    }
                    """;

            mockMvc.perform(post("/api/flowchart/generate")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content(body))
                    .andExpect(status().isOk())
                    .andExpect(jsonPath("$.svg", containsString("factorial")));
        }

        @Test
        @DisplayName("AST с if-условием → SVG содержит ДА/НЕТ")
        void svgWithDecisionContainsYesNo() throws Exception {
            String body = """
                    {
                      "ast": {
                        "type": "Program",
                        "declarations": [{
                          "type": "FunctionDecl",
                          "name": "check",
                          "returnType": {"baseType":"int","pointerLevel":0,"arraySizes":[]},
                          "parameters": [],
                          "body": {
                            "type": "BlockStmt",
                            "statements": [{
                              "type": "IfStmt",
                              "condition": {
                                "type": "BinaryExpr",
                                "operator": ">",
                                "left": {"type":"VariableExpr","name":"x","location":{"line":2,"column":8,"endLine":2,"endColumn":9}},
                                "right": {"type":"IntLiteral","value":0,"location":{"line":2,"column":12,"endLine":2,"endColumn":13}},
                                "location": {"line":2,"column":8,"endLine":2,"endColumn":13}
                              },
                              "thenBlock": {
                                "type": "ReturnStmt",
                                "value": {"type":"IntLiteral","value":1,"location":{"line":3,"column":16,"endLine":3,"endColumn":17}},
                                "location": {"line":3,"column":9,"endLine":3,"endColumn":18}
                              },
                              "location": {"line":2,"column":5,"endLine":5,"endColumn":6}
                            }],
                            "location": {"line":1,"column":16,"endLine":6,"endColumn":1}
                          },
                          "location": {"line":1,"column":1,"endLine":6,"endColumn":1}
                        }],
                        "location": {"line":1,"column":1,"endLine":6,"endColumn":1}
                      }
                    }
                    """;

            mockMvc.perform(post("/api/flowchart/generate")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content(body))
                    .andExpect(status().isOk())
                    .andExpect(jsonPath("$.svg", containsString("ДА")))
                    .andExpect(jsonPath("$.svg", containsString("НЕТ")));
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  POST /api/flowchart/generate-from-code (требует Go-сервис)
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("POST /api/flowchart/generate-from-code")
    class GenerateFromCode {

        @Test
        @DisplayName("Пустой code → 400")
        void emptyCode() throws Exception {
            mockMvc.perform(post("/api/flowchart/generate-from-code")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content("{\"code\":\"\"}"))
                    .andExpect(status().isBadRequest())
                    .andExpect(jsonPath("$.metadata.success").value(false));
        }

        @Test
        @DisplayName("Только пробелы в code → 400")
        void whitespaceCode() throws Exception {
            mockMvc.perform(post("/api/flowchart/generate-from-code")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content("{\"code\":\"   \"}"))
                    .andExpect(status().isBadRequest());
        }

        @Test
        @DisplayName("Отсутствует поле code → 400")
        void missingCodeField() throws Exception {
            mockMvc.perform(post("/api/flowchart/generate-from-code")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content("{}"))
                    .andExpect(status().isBadRequest());
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  POST /api/flowchart/generate-all-functions
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("POST /api/flowchart/generate-all-functions")
    class GenerateAllFunctions {

        @Test
        @DisplayName("Пустой code → 400")
        void emptyCode() throws Exception {
            mockMvc.perform(post("/api/flowchart/generate-all-functions")
                            .contentType(MediaType.APPLICATION_JSON)
                            .content("{\"code\":\"\"}"))
                    .andExpect(status().isBadRequest());
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  FlowchartGenerator сквозной тест (без HTTP)
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("FlowchartGenerator — сквозной тест (AST → SVG)")
    class GeneratorEndToEnd {

        @Test
        @DisplayName("Простая функция main: генерирует валидный SVG")
        void simpleMain() throws Exception {
            FlowchartGenerator gen = new FlowchartGenerator();
            String ast = """
                    {
                      "type": "Program",
                      "declarations": [{
                        "type": "FunctionDecl",
                        "name": "main",
                        "returnType": {"baseType":"int","pointerLevel":0,"arraySizes":[]},
                        "parameters": [],
                        "body": {
                          "type": "BlockStmt",
                          "statements": [{
                            "type": "ReturnStmt",
                            "value": {"type":"IntLiteral","value":0,
                                      "location":{"line":2,"column":5,"endLine":2,"endColumn":6}},
                            "location": {"line":2,"column":5,"endLine":2,"endColumn":6}
                          }],
                          "location": {"line":1,"column":14,"endLine":3,"endColumn":1}
                        },
                        "location": {"line":1,"column":1,"endLine":3,"endColumn":1}
                      }],
                      "location": {"line":1,"column":1,"endLine":3,"endColumn":1}
                    }
                    """;

            String svg = gen.generateSVG(ast);

            org.assertj.core.api.Assertions.assertThat(svg)
                    .startsWith("<?xml")
                    .contains("<svg")
                    .contains("main")
                    .contains("конец");
        }

        @Test
        @DisplayName("Функция с while-циклом: SVG содержит условие")
        void withWhileLoop() throws Exception {
            FlowchartGenerator gen = new FlowchartGenerator();
            String ast = """
                    {
                      "type": "Program",
                      "declarations": [{
                        "type": "FunctionDecl",
                        "name": "loop",
                        "returnType": {"baseType":"void","pointerLevel":0,"arraySizes":[]},
                        "parameters": [],
                        "body": {
                          "type": "BlockStmt",
                          "statements": [{
                            "type": "WhileStmt",
                            "condition": {
                              "type": "BinaryExpr",
                              "operator": "<",
                              "left": {"type":"VariableExpr","name":"i","location":{"line":2,"column":11,"endLine":2,"endColumn":12}},
                              "right": {"type":"IntLiteral","value":10,"location":{"line":2,"column":15,"endLine":2,"endColumn":17}},
                              "location": {"line":2,"column":11,"endLine":2,"endColumn":17}
                            },
                            "body": {
                              "type": "BlockStmt",
                              "statements": [],
                              "location": {"line":2,"column":19,"endLine":4,"endColumn":5}
                            },
                            "location": {"line":2,"column":5,"endLine":4,"endColumn":5}
                          }],
                          "location": {"line":1,"column":16,"endLine":5,"endColumn":1}
                        },
                        "location": {"line":1,"column":1,"endLine":5,"endColumn":1}
                      }],
                      "location": {"line":1,"column":1,"endLine":5,"endColumn":1}
                    }
                    """;

            String svg = gen.generateSVG(ast);

            org.assertj.core.api.Assertions.assertThat(svg)
                    .contains("i")
                    .contains("10");
        }

        @Test
        @DisplayName("Несколько функций: renderAllInOne содержит все имена")
        void multipleFunction() throws Exception {
            FlowchartBuilder builder = new FlowchartBuilder();
            SVGRenderer renderer = new SVGRenderer();

            // Строим AST для 2 функций
            flowchart.ast.Program p = buildTwoFunctions();
            java.util.Map<String, flowchart.model.FlowchartNode> graphs = builder.buildAllFunctions(p);
            String svg = renderer.renderAllInOne(graphs);

            org.assertj.core.api.Assertions.assertThat(svg)
                    .contains("alpha")
                    .contains("beta");
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  Хелперы
    // ─────────────────────────────────────────────────────────────

    private flowchart.ast.Program buildTwoFunctions() throws Exception {
        com.fasterxml.jackson.databind.ObjectMapper mapper =
                new com.fasterxml.jackson.databind.ObjectMapper();

        String json = """
                {
                  "type": "Program",
                  "declarations": [
                    {
                      "type": "FunctionDecl", "name": "alpha",
                      "returnType": {"baseType":"void","pointerLevel":0,"arraySizes":[]},
                      "parameters": [],
                      "body": {"type":"BlockStmt","statements":[],"location":{"line":1,"column":1,"endLine":2,"endColumn":1}},
                      "location": {"line":1,"column":1,"endLine":2,"endColumn":1}
                    },
                    {
                      "type": "FunctionDecl", "name": "beta",
                      "returnType": {"baseType":"void","pointerLevel":0,"arraySizes":[]},
                      "parameters": [],
                      "body": {"type":"BlockStmt","statements":[],"location":{"line":3,"column":1,"endLine":4,"endColumn":1}},
                      "location": {"line":3,"column":1,"endLine":4,"endColumn":1}
                    }
                  ],
                  "location": {"line":1,"column":1,"endLine":4,"endColumn":1}
                }
                """;

        return mapper.readValue(json, flowchart.ast.Program.class);
    }
}