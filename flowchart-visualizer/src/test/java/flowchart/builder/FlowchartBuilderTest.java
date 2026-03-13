package flowchart.builder;

import flowchart.AstBuilder;
import flowchart.ast.*;
import flowchart.model.*;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;

import java.util.List;
import java.util.Map;

import static flowchart.AstBuilder.*;
import static org.junit.jupiter.api.Assertions.*;

/**
 * Тесты FlowchartBuilder — проверяют структуру графа узлов,
 * построенного из AST, без проверки SVG-вывода.
 */
class FlowchartBuilderTest {

    private FlowchartBuilder builder;

    @BeforeEach
    void setUp() {
        builder = new FlowchartBuilder();
    }

    // ─────────────────────────────────────────────────────────────
    //  Базовые случаи
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("Базовая структура функции")
    class BasicFunction {

        @Test
        @DisplayName("Пустая функция: start → end")
        void emptyFunction() throws Exception {
            var program = program(func("main", "int"));
            var start = builder.buildFromProgram(program);

            assertTerminalStart(start, "main");
            var end = assertSingleNext(start);
            assertTerminalEnd(end, "конец");
        }

        @Test
        @DisplayName("Функция с одним оператором: start → process → end")
        void singleStatement() throws Exception {
            var program = program(func("main", "int",
                    varDecl("int", "x", intLit(5))
            ));
            var start = builder.buildFromProgram(program);

            assertTerminalStart(start, "main");
            var process = assertSingleNext(start);
            assertProcess(process, "int x = 5");
            var end = assertSingleNext(process);
            assertTerminalEnd(end, "конец");
        }

        @Test
        @DisplayName("Функция с несколькими операторами выстраивается в цепочку")
        void multipleStatements() throws Exception {
            var program = program(func("main", "int",
                    varDecl("int", "x", intLit(1)),
                    varDecl("int", "y", intLit(2)),
                    returnStmt(varExpr("x"))
            ));
            var start = builder.buildFromProgram(program);

            var n1 = assertSingleNext(start);
            assertProcess(n1, "int x = 1");
            var n2 = assertSingleNext(n1);
            assertProcess(n2, "int y = 2");
            // return → connector → end
            var ret = assertSingleNext(n2);
            assertProcess(ret, "return x");
        }

        @Test
        @DisplayName("Функция с параметрами: сигнатура содержит имя и тип параметра")
        void functionWithParameters() throws Exception {
            var program = program(func("factorial", "int",
                    List.<String[]>of(new String[]{"int", "n"}),
                    returnStmt(intLit(1))
            ));
            var start = builder.buildFromProgram(program);

            assertTerminalStart(start, "factorial");
            assertTrue(start.getLabel().contains("n"),
                    "Метка терминала должна содержать имя параметра 'n', но была: " + start.getLabel());
            assertTrue(start.getLabel().contains("int"),
                    "Метка терминала должна содержать тип 'int', но была: " + start.getLabel());
        }

        @Test
        @DisplayName("Ошибка если функция не найдена")
        void unknownFunctionThrows() throws Exception {
            var program = program(func("main", "int"));
            assertThrows(RuntimeException.class,
                    () -> builder.buildFromProgram(program, "nonExistent"));
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  Условные операторы
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("Условный оператор if")
    class IfStatement {

        @Test
        @DisplayName("if без else: decision с trueBranch и null falseBranch")
        void ifWithoutElse() throws Exception {
            var program = program(func("main", "int",
                    ifStmt(
                            binExpr(varExpr("x"), ">", intLit(0)),
                            varDecl("int", "y", intLit(1)),
                            null
                    )
            ));
            var start = builder.buildFromProgram(program);
            var decision = assertSingleNext(start);

            assertInstanceOf(DecisionNode.class, decision);
            var dec = (DecisionNode) decision;
            assertEquals("x > 0", dec.getLabel());
            assertNotNull(dec.getTrueBranch());
            assertProcess(dec.getTrueBranch(), "int y = 1");
            assertNull(dec.getFalseBranch());
        }

        @Test
        @DisplayName("if-else: decision с обеими ветками")
        void ifWithElse() throws Exception {
            var program = program(func("main", "int",
                    ifStmt(
                            binExpr(varExpr("a"), "<", varExpr("b")),
                            varDecl("int", "min", varExpr("a")),
                            varDecl("int", "min", varExpr("b"))
                    )
            ));
            var start = builder.buildFromProgram(program);
            var decision = assertSingleNext(start);

            assertInstanceOf(DecisionNode.class, decision);
            var dec = (DecisionNode) decision;
            assertNotNull(dec.getTrueBranch(), "Ветка ДА должна быть");
            assertNotNull(dec.getFalseBranch(), "Ветка НЕТ должна быть");
            assertProcess(dec.getTrueBranch(), "int min = a");
            assertProcess(dec.getFalseBranch(), "int min = b");
        }

        @Test
        @DisplayName("Условие сохраняется как метка")
        void conditionLabel() throws Exception {
            var program = program(func("main", "int",
                    ifStmt(
                            binExpr(varExpr("i"), "<=", intLit(10)),
                            returnStmt(intLit(1)),
                            null
                    )
            ));
            var start = builder.buildFromProgram(program);
            var decision = assertSingleNext(start);

            assertEquals("i <= 10", decision.getLabel());
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  Циклы
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("Цикл while")
    class WhileLoop {

        @Test
        @DisplayName("while создаёт LoopStartNode с телом цикла")
        void basicWhile() throws Exception {
            var program = program(func("main", "int",
                    whileStmt(
                            binExpr(varExpr("i"), "<=", intLit(5)),
                            exprStmt(assignExpr(varExpr("i"), "+=", intLit(1)))
                    )
            ));
            var start = builder.buildFromProgram(program);
            var loop = assertSingleNext(start);

            assertInstanceOf(LoopStartNode.class, loop,
                    "while должен создавать LoopStartNode");
            var loopStart = (LoopStartNode) loop;
            assertEquals("i <= 5", loopStart.getLabel());
            assertNotNull(loopStart.getLoopBody());
            assertProcess(loopStart.getLoopBody(), "i += 1");
        }

        @Test
        @DisplayName("Тело цикла связывается обратно через LoopEndNode")
        void loopBodyConnectsBack() throws Exception {
            var program = program(func("main", "int",
                    whileStmt(
                            binExpr(varExpr("i"), "<", intLit(3)),
                            exprStmt(assignExpr(varExpr("sum"), "+=", varExpr("i")))
                    )
            ));
            var start = builder.buildFromProgram(program);
            var loopStart = (LoopStartNode) assertSingleNext(start);
            var body = loopStart.getLoopBody();
            assertNotNull(body);

            // Тело должно вести к LoopEndNode, который ведёт обратно к LoopStartNode
            boolean foundLoopEnd = body.getNext().stream()
                    .anyMatch(n -> n instanceof LoopEndNode);
            assertTrue(foundLoopEnd, "Тело цикла должно содержать LoopEndNode среди next");
        }

        @Test
        @DisplayName("break внутри while создаёт ConnectorNode(break)")
        void breakInWhile() throws Exception {
            var program = program(func("main", "int",
                    whileStmt(
                            intLit(1),
                            ifStmt(
                                    binExpr(varExpr("x"), ">", intLit(10)),
                                    breakStmt(),
                                    null
                            )
                    )
            ));
            var start = builder.buildFromProgram(program);
            var loop = assertSingleNext(start);
            assertInstanceOf(LoopStartNode.class, loop);

            // Ищем ConnectorNode("break") в теле
            assertTrue(
                    containsConnector(loop, "break"),
                    "Граф должен содержать ConnectorNode(break)"
            );
        }

        @Test
        @DisplayName("continue внутри while создаёт ConnectorNode(continue)")
        void continueInWhile() throws Exception {
            var program = program(func("main", "int",
                    whileStmt(
                            binExpr(varExpr("i"), "<", intLit(10)),
                            ifStmt(
                                    binExpr(varExpr("i"), "%", intLit(2)),
                                    continueStmt(),
                                    null
                            )
                    )
            ));
            var start = builder.buildFromProgram(program);
            assertTrue(
                    containsConnector(start, "continue"),
                    "Граф должен содержать ConnectorNode(continue)"
            );
        }
    }

    @Nested
    @DisplayName("Цикл for")
    class ForLoop {

        @Test
        @DisplayName("for создаёт: init → LoopStart → body")
        void basicFor() throws Exception {
            var program = program(func("main", "int",
                    forStmt(
                            varDecl("int", "i", intLit(0)),
                            binExpr(varExpr("i"), "<", intLit(5)),
                            exprStmt(unaryExpr(varExpr("i"), "++", true)),
                            exprStmt(assignExpr(varExpr("sum"), "+=", varExpr("i")))
                    )
            ));
            var start = builder.buildFromProgram(program);

            // первый узел после start — init
            var init = assertSingleNext(start);
            assertProcess(init, "int i = 0");

            // следующий — LoopStartNode
            var loop = assertSingleNext(init);
            assertInstanceOf(LoopStartNode.class, loop);
            assertEquals("i < 5", loop.getLabel());
        }

        @Test
        @DisplayName("for без init: сразу LoopStartNode")
        void forWithoutInit() throws Exception {
            var program = program(func("main", "int",
                    forStmt(
                            null,
                            binExpr(varExpr("i"), "<", intLit(5)),
                            null,
                            returnStmt(intLit(0))
                    )
            ));
            var start = builder.buildFromProgram(program);
            var loop = assertSingleNext(start);
            assertInstanceOf(LoopStartNode.class, loop, "Без init первым должен быть LoopStartNode");
        }
    }

    @Nested
    @DisplayName("Цикл do-while")
    class DoWhileLoop {

        @Test
        @DisplayName("do-while создаёт DoWhileNode с телом")
        void basicDoWhile() throws Exception {
            var program = program(func("main", "void",
                    doWhileStmt(
                            binExpr(varExpr("year"), "<=", intLit(2040)),
                            exprStmt(assignExpr(varExpr("year"), "+=", intLit(1)))
                    )
            ));
            var start = builder.buildFromProgram(program);
            var doWhile = assertSingleNext(start);

            assertInstanceOf(DoWhileNode.class, doWhile,
                    "do-while должен создавать DoWhileNode");
            var dwn = (DoWhileNode) doWhile;
            assertEquals("year <= 2040", dwn.getLabel());
            assertNotNull(dwn.getLoopBody());
            assertProcess(dwn.getLoopBody(), "year += 1");
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  Return
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("Оператор return")
    class ReturnStatement {

        @Test
        @DisplayName("return с значением: ProcessNode(return X) → ConnectorNode(return)")
        void returnWithValue() throws Exception {
            var program = program(func("main", "int",
                    returnStmt(intLit(42))
            ));
            var start = builder.buildFromProgram(program);
            var process = assertSingleNext(start);
            assertProcess(process, "return 42");

            var connector = assertSingleNext(process);
            assertInstanceOf(ConnectorNode.class, connector);
            assertEquals("return", connector.getLabel());
        }

        @Test
        @DisplayName("return без значения: ConnectorNode(return)")
        void returnWithoutValue() throws Exception {
            var program = program(func("main", "void",
                    returnStmt(null)
            ));
            var start = builder.buildFromProgram(program);
            var connector = assertSingleNext(start);
            assertInstanceOf(ConnectorNode.class, connector);
            assertEquals("return", connector.getLabel());
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  Несколько функций
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("Несколько функций")
    class MultipleFunctions {

        @Test
        @DisplayName("getFunctionNames возвращает все имена")
        void getFunctionNames() throws Exception {
            var program = program(
                    func("isPrime", "int", List.<String[]>of(new String[]{"int", "num"})),
                    func("main", "int")
            );
            var names = builder.getFunctionNames(program);
            assertEquals(2, names.size());
            assertTrue(names.contains("isPrime"));
            assertTrue(names.contains("main"));
        }

        @Test
        @DisplayName("buildAllFunctions возвращает граф для каждой функции")
        void buildAllFunctions() throws Exception {
            var program = program(
                    func("helper", "int", varDecl("int", "x", intLit(1))),
                    func("main", "int", varDecl("int", "y", intLit(2)))
            );
            Map<String, FlowchartNode> all = builder.buildAllFunctions(program);

            assertEquals(2, all.size());
            assertTrue(all.containsKey("helper"));
            assertTrue(all.containsKey("main"));
            assertTerminalStart(all.get("helper"), "helper");
            assertTerminalStart(all.get("main"), "main");
        }

        @Test
        @DisplayName("buildFromProgram('main') строит только main")
        void buildSpecificFunction() throws Exception {
            var program = program(
                    func("helper", "int"),
                    func("main", "int", varDecl("int", "x", intLit(0)))
            );
            var start = builder.buildFromProgram(program, "main");
            assertTerminalStart(start, "main");
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  Выражения
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("Текст выражений в узлах")
    class ExpressionLabels {

        @Test
        @DisplayName("Бинарное выражение: a + b")
        void binaryExpr() throws Exception {
            var program = program(func("main", "int",
                    varDecl("int", "s", binExpr(varExpr("a"), "+", varExpr("b")))
            ));
            var start = builder.buildFromProgram(program);
            var process = assertSingleNext(start);
            assertEquals("int s = a + b", process.getLabel());
        }

        @Test
        @DisplayName("Постфиксный инкремент: i++")
        void postfixIncrement() throws Exception {
            var program = program(func("main", "int",
                    exprStmt(unaryExpr(varExpr("i"), "++", true))
            ));
            var start = builder.buildFromProgram(program);
            var process = assertSingleNext(start);
            assertEquals("++i", process.getLabel());
        }

        @Test
        @DisplayName("Присваивание: sum += i")
        void compoundAssign() throws Exception {
            var program = program(func("main", "int",
                    exprStmt(assignExpr(varExpr("sum"), "+=", varExpr("i")))
            ));
            var start = builder.buildFromProgram(program);
            var process = assertSingleNext(start);
            assertEquals("sum += i", process.getLabel());
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  Вспомогательные assert-методы
    // ─────────────────────────────────────────────────────────────

    private void assertTerminalStart(FlowchartNode node, String name) {
        assertInstanceOf(TerminalNode.class, node);
        assertTrue(((TerminalNode) node).isStart());
        assertTrue(node.getLabel().startsWith(name),
                "Ожидалась метка, начинающаяся с '%s', но была: '%s'".formatted(name, node.getLabel()));
    }

    private void assertTerminalEnd(FlowchartNode node, String label) {
        assertInstanceOf(TerminalNode.class, node);
        assertFalse(((TerminalNode) node).isStart());
        assertEquals(label, node.getLabel());
    }

    private void assertProcess(FlowchartNode node, String label) {
        assertInstanceOf(ProcessNode.class, node,
                "Ожидался ProcessNode с меткой '%s', но был %s".formatted(label, node.getClass().getSimpleName()));
        assertEquals(label, node.getLabel());
    }

    private FlowchartNode assertSingleNext(FlowchartNode node) {
        assertFalse(node.getNext().isEmpty(),
                "Узел '%s' не имеет следующих узлов".formatted(node.getLabel()));
        return node.getNext().get(0);
    }

    /** Рекурсивно ищет ConnectorNode с заданной меткой */
    private boolean containsConnector(FlowchartNode node, String label) {
        return containsConnector(node, label, new java.util.HashSet<>());
    }

    private boolean containsConnector(FlowchartNode node, String label, java.util.Set<FlowchartNode> visited) {
        if (node == null || visited.contains(node)) return false;
        visited.add(node);
        if (node instanceof ConnectorNode && label.equals(node.getLabel())) return true;
        if (node instanceof DecisionNode dn) {
            if (containsConnector(dn.getTrueBranch(), label, visited)) return true;
            if (containsConnector(dn.getFalseBranch(), label, visited)) return true;
        }
        if (node instanceof LoopStartNode ls) {
            if (containsConnector(ls.getLoopBody(), label, visited)) return true;
        }
        for (var next : node.getNext()) {
            if (containsConnector(next, label, visited)) return true;
        }
        return false;
    }
}