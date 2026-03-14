package flowchart.renderer;

import flowchart.AstBuilder;
import flowchart.builder.FlowchartBuilder;
import flowchart.model.*;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;

import java.util.LinkedHashMap;
import java.util.Map;

import java.util.List;
import static flowchart.AstBuilder.*;
import static org.junit.jupiter.api.Assertions.*;

/**
 * Тесты SVGRenderer — проверяют корректность SVG-вывода:
 * наличие нужных элементов, структуру XML, содержимое меток.
 */
class SVGRendererTest {

    private SVGRenderer renderer;
    private FlowchartBuilder builder;

    @BeforeEach
    void setUp() {
        renderer = new SVGRenderer();
        builder  = new FlowchartBuilder();
    }

    // ─────────────────────────────────────────────────────────────
    //  Структура SVG-документа
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("Структура SVG-документа")
    class SvgStructure {

        @Test
        @DisplayName("SVG начинается с XML-заголовка")
        void hasXmlHeader() throws Exception {
            var svg = renderMain();
            assertTrue(svg.startsWith("<?xml"), "SVG должен начинаться с XML-заголовка");
        }

        @Test
        @DisplayName("SVG содержит корневой тег <svg>")
        void hasSvgTag() throws Exception {
            var svg = renderMain();
            assertTrue(svg.contains("<svg ") && svg.contains("</svg>"),
                    "SVG должен содержать открывающий и закрывающий теги <svg>");
        }

        @Test
        @DisplayName("SVG содержит viewBox")
        void hasViewBox() throws Exception {
            var svg = renderMain();
            assertTrue(svg.contains("viewBox="), "SVG должен содержать атрибут viewBox");
        }

        @Test
        @DisplayName("SVG содержит определение стрелок-маркеров")
        void hasArrowMarker() throws Exception {
            var svg = renderMain();
            assertTrue(svg.contains("id=\"arrow\""), "SVG должен содержать маркер стрелки");
        }

        @Test
        @DisplayName("SVG содержит CSS-стили")
        void hasCssStyles() throws Exception {
            var svg = renderMain();
            assertTrue(svg.contains("<style>") && svg.contains("</style>"),
                    "SVG должен содержать блок стилей");
            assertTrue(svg.contains(".shape"), "Стили должны включать .shape");
            assertTrue(svg.contains(".text"),  "Стили должны включать .text");
        }

        @Test
        @DisplayName("SVG валиден — все теги закрыты")
        void tagsAreClosed() throws Exception {
            var svg = renderMain();
            // Простая проверка: каждый открытый составной тег имеет закрывающий
            assertTrue(svg.contains("</svg>"), "Не хватает </svg>");
            assertTrue(svg.contains("</defs>"), "Не хватает </defs>");
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  Терминальные блоки (овалы)
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("Терминальные блоки")
    class TerminalBlocks {

        @Test
        @DisplayName("Пустая функция: содержит эллипс начала и конца")
        void emptyFunctionHasTwoEllipses() throws Exception {
            var svg = renderMain();
            int count = countOccurrences(svg, "<ellipse");
            assertEquals(2, count, "Пустая функция должна содержать ровно 2 эллипса (начало и конец)");
        }

        @Test
        @DisplayName("Метка начального терминала присутствует в SVG")
        void startLabelInSvg() throws Exception {
            var program = program(func("factorial", "int",
                    List.<String[]>of(new String[]{"int", "n"})
            ));
            var start = builder.buildFromProgram(program);
            var svg = renderer.render(start);

            assertTrue(svg.contains("factorial"), "SVG должен содержать имя функции 'factorial'");
        }

        @Test
        @DisplayName("Метка конечного терминала 'конец' присутствует в SVG")
        void endLabelInSvg() throws Exception {
            var svg = renderMain();
            assertTrue(svg.contains("конец"), "SVG должен содержать метку завершения 'конец'");
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  Блоки процессов (прямоугольники)
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("Блоки процессов")
    class ProcessBlocks {

        @Test
        @DisplayName("Один оператор → один прямоугольник")
        void oneStatementOneRect() throws Exception {
            var program = program(func("main", "int",
                    varDecl("int", "x", intLit(5))
            ));
            var start = builder.buildFromProgram(program);
            var svg = renderer.render(start);

            int count = countOccurrences(svg, "<rect");
            assertEquals(1, count, "Один оператор → один прямоугольник");
        }

        @Test
        @DisplayName("Текст оператора присутствует в SVG")
        void operatorTextInSvg() throws Exception {
            var program = program(func("main", "int",
                    varDecl("int", "sum", intLit(0))
            ));
            var start = builder.buildFromProgram(program);
            var svg = renderer.render(start);

            assertTrue(svg.contains("int sum = 0"),
                    "SVG должен содержать текст оператора");
        }

        @Test
        @DisplayName("Несколько операторов → несколько прямоугольников")
        void multipleStatementsMultipleRects() throws Exception {
            var program = program(func("main", "int",
                    varDecl("int", "a", intLit(1)),
                    varDecl("int", "b", intLit(2)),
                    varDecl("int", "c", intLit(3))
            ));
            var start = builder.buildFromProgram(program);
            var svg = renderer.render(start);

            int count = countOccurrences(svg, "<rect");
            assertEquals(3, count, "Три оператора → три прямоугольника");
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  Ромбы решений
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("Блоки решений (ромбы)")
    class DecisionBlocks {

        @Test
        @DisplayName("if → один ромб в SVG")
        void ifOnePolygon() throws Exception {
            var program = program(func("main", "int",
                    ifStmt(
                            binExpr(varExpr("x"), ">", intLit(0)),
                            varDecl("int", "y", intLit(1)),
                            null
                    )
            ));
            var start = builder.buildFromProgram(program);
            var svg = renderer.render(start);

            assertTrue(countOccurrences(svg, "<polygon") >= 1,
                    "if → должен быть хотя бы один ромб (<polygon>)");
        }

        @Test
        @DisplayName("Условие if присутствует в SVG")
        void ifConditionInSvg() throws Exception {
            var program = program(func("main", "int",
                    ifStmt(
                            binExpr(varExpr("i"), "<=", intLit(10)),
                            returnStmt(intLit(1)),
                            null
                    )
            ));
            var start = builder.buildFromProgram(program);
            var svg = renderer.render(start);

            assertTrue(svg.contains("i &lt;= 10") || svg.contains("i <= 10"),
                    "SVG должен содержать условие (с экранированием или без)");
        }

        @Test
        @DisplayName("Метки ДА/НЕТ присутствуют в SVG")
        void yesNoLabels() throws Exception {
            var program = program(func("main", "int",
                    ifStmt(
                            binExpr(varExpr("x"), ">", intLit(0)),
                            varDecl("int", "y", intLit(1)),
                            null
                    )
            ));
            var start = builder.buildFromProgram(program);
            var svg = renderer.render(start);

            assertTrue(svg.contains("ДА"),  "SVG должен содержать метку 'ДА'");
            assertTrue(svg.contains("НЕТ"), "SVG должен содержать метку 'НЕТ'");
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  Циклы
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("Циклы в SVG")
    class LoopRendering {

        @Test
        @DisplayName("while → ромб условия с меткой ДА/НЕТ")
        void whileLoopHasDiamondAndLabels() throws Exception {
            var program = program(func("main", "int",
                    whileStmt(
                            binExpr(varExpr("i"), "<=", intLit(5)),
                            exprStmt(assignExpr(varExpr("i"), "+=", intLit(1)))
                    )
            ));
            var start = builder.buildFromProgram(program);
            var svg = renderer.render(start);

            assertTrue(svg.contains("i &lt;= 5") || svg.contains("i <= 5"),
                    "SVG должен содержать условие while");
            assertTrue(svg.contains("ДА"), "SVG должен содержать 'ДА'");
            assertTrue(svg.contains("НЕТ"), "SVG должен содержать 'НЕТ'");
        }

        @Test
        @DisplayName("for → инициализирующий прямоугольник и ромб условия")
        void forLoopHasInitAndCondition() throws Exception {
            var program = program(func("main", "int",
                    forStmt(
                            varDecl("int", "i", intLit(0)),
                            binExpr(varExpr("i"), "<", intLit(10)),
                            exprStmt(unaryExpr(varExpr("i"), "++", true)),
                            exprStmt(assignExpr(varExpr("sum"), "+=", varExpr("i")))
                    )
            ));
            var start = builder.buildFromProgram(program);
            var svg = renderer.render(start);

            assertTrue(svg.contains("int i = 0"), "SVG должен содержать инициализацию 'int i = 0'");
            assertTrue(svg.contains("i &lt; 10") || svg.contains("i < 10"),
                    "SVG должен содержать условие 'i < 10'");
        }

        @Test
        @DisplayName("do-while → условие внизу, тело выше")
        void doWhileConditionBelowBody() throws Exception {
            var program = program(func("main", "void",
                    doWhileStmt(
                            binExpr(varExpr("year"), "<=", intLit(2040)),
                            exprStmt(assignExpr(varExpr("year"), "+=", intLit(1)))
                    )
            ));
            var start = builder.buildFromProgram(program);
            var svg = renderer.render(start);

            int yearPlusIdx = svg.indexOf("year += 1");
            int condIdx = svg.indexOf("year &lt;= 2040");
            if (condIdx < 0) condIdx = svg.indexOf("year <= 2040");

            assertTrue(yearPlusIdx >= 0, "SVG должен содержать тело цикла 'year += 1'");
            assertTrue(condIdx >= 0, "SVG должен содержать условие do-while");
            assertTrue(yearPlusIdx < condIdx,
                    "Тело do-while должно рендериться до условия (выше по Y)");
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  renderAll и renderAllInOne
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("Несколько функций")
    class MultipleFunctionsRendering {

        @Test
        @DisplayName("renderAll возвращает SVG для каждой функции")
        void renderAllReturnsSvgPerFunction() throws Exception {
            var program = program(
                    func("helper", "int", varDecl("int", "x", intLit(1))),
                    func("main", "int",   varDecl("int", "y", intLit(2)))
            );
            var graphs = builder.buildAllFunctions(program);
            var svgs   = renderer.renderAll(graphs);

            assertEquals(2, svgs.size());
            assertTrue(svgs.containsKey("helper"));
            assertTrue(svgs.containsKey("main"));
            assertTrue(svgs.get("helper").contains("helper"));
            assertTrue(svgs.get("main").contains("main"));
        }

        @Test
        @DisplayName("renderAllInOne возвращает один SVG с обоими именами функций")
        void renderAllInOneContainsBothNames() throws Exception {
            var program = program(
                    func("isPrime", "int", List.<String[]>of(new String[]{"int", "num"})),
                    func("main", "int")
            );
            var graphs = builder.buildAllFunctions(program);
            var svg    = renderer.renderAllInOne(graphs);

            assertTrue(svg.contains("isPrime"), "Общий SVG должен содержать 'isPrime'");
            assertTrue(svg.contains("main"),    "Общий SVG должен содержать 'main'");
        }

        @Test
        @DisplayName("renderAllInOne: main идёт первым (сортировка)")
        void mainIsFirstInAllInOne() throws Exception {
            var program = program(
                    func("zzz", "void"),
                    func("main", "int")
            );
            var graphs = builder.buildAllFunctions(program);
            var svg    = renderer.renderAllInOne(graphs);

            int mainIdx = svg.indexOf(">main<");
            int zzzIdx  = svg.indexOf(">zzz<");
            assertTrue(mainIdx >= 0 && zzzIdx >= 0, "Оба имени должны быть в SVG");
            assertTrue(mainIdx < zzzIdx, "'main' должна идти раньше 'zzz' в SVG");
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  Экранирование XML
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("Экранирование XML")
    class XmlEscaping {

        @Test
        @DisplayName("Символ < экранируется в &lt;")
        void lessThanEscaped() throws Exception {
            var program = program(func("main", "int",
                    ifStmt(
                            binExpr(varExpr("a"), "<", intLit(5)),
                            returnStmt(intLit(1)),
                            null
                    )
            ));
            var start = builder.buildFromProgram(program);
            var svg = renderer.render(start);

            assertFalse(svg.contains("a < 5"),
                    "Символ < не должен присутствовать незаэкранированным внутри текстового узла");
            assertTrue(svg.contains("a &lt; 5"),
                    "Символ < должен быть заэкранирован как &lt;");
        }

        @Test
        @DisplayName("Символ > экранируется в &gt;")
        void greaterThanEscaped() throws Exception {
            var program = program(func("main", "int",
                    ifStmt(
                            binExpr(varExpr("x"), ">", intLit(0)),
                            returnStmt(intLit(1)),
                            null
                    )
            ));
            var start = builder.buildFromProgram(program);
            var svg = renderer.render(start);

            assertTrue(svg.contains("&gt;"),
                    "Символ > должен быть заэкранирован как &gt;");
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  FlowchartGenerator (интеграция builder + renderer)
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("FlowchartGenerator — интеграция")
    class GeneratorIntegration {

        @Test
        @DisplayName("generateSVG из JSON AST возвращает валидный SVG")
        void generateFromJson() throws Exception {
            var gen = new flowchart.FlowchartGenerator();
            String json = """
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
            var svg = gen.generateSVG(json);
            assertTrue(svg.startsWith("<?xml"), "generateSVG должен возвращать XML");
            assertTrue(svg.contains("<svg"),    "generateSVG должен содержать <svg>");
            assertTrue(svg.contains("main"),    "SVG должен содержать имя функции");
        }

        @Test
        @DisplayName("generateSVG для факториала с параметром (int n) содержит параметр в метке")
        void generateFactorialWithParam() throws Exception {
            var gen = new flowchart.FlowchartGenerator();
            String json = """
                {
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
                      "statements": [{
                        "type": "ReturnStmt",
                        "value": {"type":"IntLiteral","value":1,
                                  "location":{"line":2,"column":5,"endLine":2,"endColumn":6}},
                        "location": {"line":2,"column":5,"endLine":2,"endColumn":6}
                      }],
                      "location": {"line":1,"column":18,"endLine":3,"endColumn":1}
                    },
                    "location": {"line":1,"column":1,"endLine":3,"endColumn":1}
                  }],
                  "location": {"line":1,"column":1,"endLine":3,"endColumn":1}
                }
                """;
            var svg = gen.generateSVG(json);

            assertTrue(svg.contains("factorial"), "SVG должен содержать 'factorial'");
            assertTrue(svg.contains("n"),         "SVG должен содержать параметр 'n'");
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  Вспомогательные методы
    // ─────────────────────────────────────────────────────────────

    /** Рендерит функцию main без операторов */
    private String renderMain() throws Exception {
        var program = program(func("main", "int"));
        var start   = builder.buildFromProgram(program);
        return renderer.render(start);
    }

    private int countOccurrences(String text, String pattern) {
        int count = 0, idx = 0;
        while ((idx = text.indexOf(pattern, idx)) != -1) {
            count++;
            idx += pattern.length();
        }
        return count;
    }
}