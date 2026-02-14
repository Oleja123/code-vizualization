package flowchart;

import com.fasterxml.jackson.databind.ObjectMapper;
import flowchart.ast.Program;
import flowchart.builder.FlowchartBuilder;
import flowchart.model.FlowchartNode;
import flowchart.renderer.SVGRenderer;


import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Paths;

/**
 * Главный API для генерации блок-схем из AST
 * 
 * Использование:
 * 1. Получить JSON AST от Go сервиса
 * 2. FlowchartGenerator.generateSVG(jsonString)
 * 3. Вернуть SVG клиенту для отображения в браузере
 */
public class FlowchartGenerator {
    private final ObjectMapper objectMapper;
    private final FlowchartBuilder builder;
    private final SVGRenderer renderer;
    
    public FlowchartGenerator() {
        this.objectMapper = new ObjectMapper();
        this.builder = new FlowchartBuilder();
        this.renderer = new SVGRenderer();
    }
    
    /**
     * Генерация SVG блок-схемы из JSON AST
     * 
     * @param astJson JSON строка с AST от Go сервиса
     * @return SVG строка для отображения в браузере
     */
    public String generateSVG(String astJson) throws IOException {
        // 1. Парсим JSON в объектную модель
        Program program = objectMapper.readValue(astJson, Program.class);
        
        // 2. Строим граф блок-схемы
        FlowchartNode flowchart = builder.buildFromProgram(program);
        
        // 3. Рендерим в SVG
        return renderer.render(flowchart);
    }
    
    /**
     * Генерация SVG из файла с AST
     */
    public String generateSVGFromFile(String astFilePath) throws IOException {
        String astJson = new String(Files.readAllBytes(Paths.get(astFilePath)));
        return generateSVG(astJson);
    }
    
    /**
     * Сохранить SVG в файл
     */
    public void generateSVGToFile(String astJson, String outputPath) throws IOException {
        String svg = generateSVG(astJson);
        Files.write(Paths.get(outputPath), svg.getBytes());
    }
    
    /**
     * Пример использования
     */
    public static void main(String[] args) {
        try {
            FlowchartGenerator generator = new FlowchartGenerator();
            
            // Пример: простая программа
            String exampleAST = """
                {
                    "type": "Program",
                    "declarations": [
                        {
                            "type": "FunctionDecl",
                            "name": "main",
                            "returnType": {
                                "baseType": "int",
                                "pointerLevel": 0,
                                "arraySizes": []
                            },
                            "parameters": [],
                            "body": {
                                "type": "BlockStmt",
                                "statements": [
                                    {
                                        "type": "VariableDecl",
                                        "varType": {
                                            "baseType": "int",
                                            "pointerLevel": 0,
                                            "arraySizes": []
                                        },
                                        "name": "x",
                                        "initExpr": {
                                            "type": "IntLiteral",
                                            "value": 10,
                                            "location": {"line": 2, "column": 13, "endLine": 2, "endColumn": 15}
                                        },
                                        "location": {"line": 2, "column": 5, "endLine": 2, "endColumn": 16}
                                    },
                                    {
                                        "type": "IfStmt",
                                        "condition": {
                                            "type": "BinaryExpr",
                                            "op": ">",
                                            "left": {
                                                "type": "VariableExpr",
                                                "name": "x",
                                                "location": {"line": 3, "column": 9, "endLine": 3, "endColumn": 10}
                                            },
                                            "right": {
                                                "type": "IntLiteral",
                                                "value": 5,
                                                "location": {"line": 3, "column": 13, "endLine": 3, "endColumn": 14}
                                            },
                                            "location": {"line": 3, "column": 9, "endLine": 3, "endColumn": 14}
                                        },
                                        "thenBlock": {
                                            "type": "ReturnStmt",
                                            "value": {
                                                "type": "IntLiteral",
                                                "value": 1,
                                                "location": {"line": 4, "column": 16, "endLine": 4, "endColumn": 17}
                                            },
                                            "location": {"line": 4, "column": 9, "endLine": 4, "endColumn": 18}
                                        },
                                        "location": {"line": 3, "column": 5, "endLine": 5, "endColumn": 6}
                                    },
                                    {
                                        "type": "ReturnStmt",
                                        "value": {
                                            "type": "IntLiteral",
                                            "value": 0,
                                            "location": {"line": 6, "column": 12, "endLine": 6, "endColumn": 13}
                                        },
                                        "location": {"line": 6, "column": 5, "endLine": 6, "endColumn": 14}
                                    }
                                ],
                                "location": {"line": 1, "column": 16, "endLine": 7, "endColumn": 2}
                            },
                            "location": {"line": 1, "column": 1, "endLine": 7, "endColumn": 2}
                        }
                    ],
                    "location": {"line": 1, "column": 1, "endLine": 7, "endColumn": 2}
                }
                """;
            
            // Генерация SVG
            String svg = generator.generateSVG(exampleAST);
            
            // Сохранение в файл
            generator.generateSVGToFile(exampleAST, "output.svg");
            
            System.out.println("✓ Блок-схема успешно создана!");
            System.out.println("SVG длина: " + svg.length() + " символов");
            
        } catch (Exception e) {
            System.err.println("Ошибка: " + e.getMessage());
            e.printStackTrace();
        }
    }
}
