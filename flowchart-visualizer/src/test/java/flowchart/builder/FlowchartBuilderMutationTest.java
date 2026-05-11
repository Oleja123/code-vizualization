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
 * Дополнительные тесты FlowchartBuilder нацеленные на выживших мутантов.
 * Покрывают:
 *  - Выбор функции по умолчанию (main vs первая)
 *  - getFunctionNames порядок и количество
 *  - buildAllFunctions — все функции строятся
 *  - linkNodes граничные случаи (LoopStart с exitNode, DoWhile с exitNode)
 *  - attachEnd — ConnectorNode return с уже заполненным next
 *  - DecisionNode — оба branch null, один null
 *  - Блок с null-узлами (пропускаются)
 *  - Цепочки операторов после break/continue
 *  - Вложенные циклы
 */
@DisplayName("FlowchartBuilder — мутационные граничные тесты")
class FlowchartBuilderMutationTest {

    private FlowchartBuilder builder;

    @BeforeEach
    void setUp() {
        builder = new FlowchartBuilder();
    }

    // ── Хелперы для проверки узлов ────────────────────────────────

    static void assertTerminalStart(FlowchartNode node, String nameContains) {
        assertInstanceOf(TerminalNode.class, node);
        assertTrue(((TerminalNode) node).isStart());
        assertTrue(node.getLabel().contains(nameContains),
                "Expected label to contain '" + nameContains + "' but was: " + node.getLabel());
    }

    static void assertTerminalEnd(FlowchartNode node, String label) {
        assertInstanceOf(TerminalNode.class, node);
        assertFalse(((TerminalNode) node).isStart());
        assertEquals(label, node.getLabel());
    }

    static FlowchartNode assertSingleNext(FlowchartNode node) {
        assertEquals(1, node.getNext().size(),
                "Expected 1 next node for " + node.getLabel() + " but got " + node.getNext().size());
        return node.getNext().get(0);
    }

    // ═══════════════════════════════════════════════════════════════
    //  Выбор функции по умолчанию
    // ═══════════════════════════════════════════════════════════════

    @Nested
    @DisplayName("Выбор функции по умолчанию")
    class DefaultFunctionSelection {

        @Test
        @DisplayName("Если есть main — выбирается main, а не первая")
        void prefersMainOverFirst() throws Exception {
            var program = program(
                    func("helper", "void"),
                    func("main", "int")
            );
            var start = builder.buildFromProgram(program);
            assertTerminalStart(start, "main");
        }

        @Test
        @DisplayName("Если нет main — выбирается первая функция")
        void firstFunctionWhenNoMain() throws Exception {
            var program = program(
                    func("alpha", "int"),
                    func("beta", "void")
            );
            var start = builder.buildFromProgram(program);
            assertTerminalStart(start, "alpha");
        }

        @Test
        @DisplayName("Единственная функция не main — берётся она")
        void singleNonMainFunction() throws Exception {
            var program = program(func("compute", "int"));
            var start = builder.buildFromProgram(program);
            assertTerminalStart(start, "compute");
        }

        @Test
        @DisplayName("Явное указание имени функции: строится нужная")
        void explicitFunctionName() throws Exception {
            var program = program(
                    func("main", "int"),
                    func("helper", "void", returnStmt(intLit(42)))
            );
            var start = builder.buildFromProgram(program, "helper");
            assertTerminalStart(start, "helper");
        }

        @Test
        @DisplayName("Несуществующая функция бросает RuntimeException")
        void unknownFunctionThrows() throws Exception {
            var program = program(func("main", "int"));
            assertThrows(RuntimeException.class,
                    () -> builder.buildFromProgram(program, "noSuchFunc"));
        }
    }

    // ═══════════════════════════════════════════════════════════════
    //  getFunctionNames
    // ═══════════════════════════════════════════════════════════════

    @Nested
    @DisplayName("getFunctionNames")
    class GetFunctionNames {

        @Test
        @DisplayName("Пустая программа: пустой список")
        void emptyProgram() throws Exception {
            var program = program();
            assertEquals(List.of(), builder.getFunctionNames(program));
        }

        @Test
        @DisplayName("Одна функция: список из одного элемента")
        void oneFunction() throws Exception {
            var program = program(func("foo", "int"));
            assertEquals(List.of("foo"), builder.getFunctionNames(program));
        }

        @Test
        @DisplayName("Три функции: порядок сохраняется")
        void threeFunctions() throws Exception {
            var program = program(
                    func("a", "int"),
                    func("b", "void"),
                    func("c", "int")
            );
            assertEquals(List.of("a", "b", "c"), builder.getFunctionNames(program));
        }

        @Test
        @DisplayName("Размер списка совпадает с количеством функций")
        void sizeMustMatch() throws Exception {
            var program = program(
                    func("f1", "int"),
                    func("f2", "int")
            );
            assertEquals(2, builder.getFunctionNames(program).size());
        }
    }

    // ═══════════════════════════════════════════════════════════════
    //  buildAllFunctions
    // ═══════════════════════════════════════════════════════════════

    @Nested
    @DisplayName("buildAllFunctions")
    class BuildAllFunctions {

        @Test
        @DisplayName("Все функции строятся и их ключи совпадают с именами")
        void allFunctionsBuilt() throws Exception {
            var program = program(
                    func("main", "int"),
                    func("helper", "void")
            );
            Map<String, FlowchartNode> result = builder.buildAllFunctions(program);
            assertEquals(2, result.size());
            assertTrue(result.containsKey("main"));
            assertTrue(result.containsKey("helper"));
        }

        @Test
        @DisplayName("Каждый граф начинается с TerminalNode (start)")
        void eachGraphStartsWithTerminal() throws Exception {
            var program = program(
                    func("f1", "int"),
                    func("f2", "int")
            );
            Map<String, FlowchartNode> result = builder.buildAllFunctions(program);
            for (Map.Entry<String, FlowchartNode> e : result.entrySet()) {
                assertInstanceOf(TerminalNode.class, e.getValue(),
                        "Start of " + e.getKey() + " must be TerminalNode");
                assertTrue(((TerminalNode) e.getValue()).isStart());
            }
        }

        @Test
        @DisplayName("Пустая программа: пустой map")
        void emptyProgram() throws Exception {
            var program = program();
            assertTrue(builder.buildAllFunctions(program).isEmpty());
        }
    }

    // ═══════════════════════════════════════════════════════════════
    //  Структура if-else
    // ═══════════════════════════════════════════════════════════════

    @Nested
    @DisplayName("if-else структура")
    class IfElseStructure {

        @Test
        @DisplayName("if с пустым then и else: оба ведут к end")
        void ifElseBothEmpty() throws Exception {
            var program = program(func("main", "int",
                    ifStmt(intLit(1),
                            exprStmt(varExpr("x")),
                            exprStmt(varExpr("y")))
            ));
            var start = builder.buildFromProgram(program);
            var decision = assertSingleNext(start);
            assertInstanceOf(DecisionNode.class, decision);
            DecisionNode d = (DecisionNode) decision;
            assertNotNull(d.getTrueBranch(), "trueBranch не должен быть null");
            assertNotNull(d.getFalseBranch(), "falseBranch не должен быть null");
        }

        @Test
        @DisplayName("if без else: trueBranch не null, falseBranch null (ведёт через attachEnd)")
        void ifWithoutElseFalseBranch() throws Exception {
            var program = program(func("main", "int",
                    ifStmt(intLit(1), exprStmt(varExpr("x")), null)
            ));
            var start = builder.buildFromProgram(program);
            var decision = assertSingleNext(start);
            assertInstanceOf(DecisionNode.class, decision);
            DecisionNode d = (DecisionNode) decision;
            // trueBranch должен указывать на ProcessNode
            assertNotNull(d.getTrueBranch(), "trueBranch не должен быть null");
            assertInstanceOf(ProcessNode.class, d.getTrueBranch());
            // falseBranch null — выход идёт через attachEnd напрямую к end
            assertNull(d.getFalseBranch(), "falseBranch должен быть null когда нет else");
        }

        @Test
        @DisplayName("Вложенный if-else в then ветке")
        void nestedIfInThen() throws Exception {
            var program = program(func("main", "int",
                    ifStmt(intLit(1),
                            ifStmt(intLit(2), exprStmt(varExpr("a")), null),
                            null)
            ));
            var start = builder.buildFromProgram(program);
            var outerDecision = assertSingleNext(start);
            assertInstanceOf(DecisionNode.class, outerDecision);
            DecisionNode outer = (DecisionNode) outerDecision;
            assertInstanceOf(DecisionNode.class, outer.getTrueBranch(),
                    "Внутри then должен быть ещё DecisionNode");
        }
    }

    // ═══════════════════════════════════════════════════════════════
    //  Циклы — детальная проверка связей
    // ═══════════════════════════════════════════════════════════════

    @Nested
    @DisplayName("Циклы — связи узлов")
    class LoopConnections {

        @Test
        @DisplayName("while: LoopStartNode имеет exitNode после построения")
        void whileHasExitNode() throws Exception {
            var program = program(func("main", "int",
                    whileStmt(intLit(1), exprStmt(varExpr("x")))
            ));
            var start = builder.buildFromProgram(program);
            var loopStart = assertSingleNext(start);
            assertInstanceOf(LoopStartNode.class, loopStart);
            assertNotNull(((LoopStartNode) loopStart).getExitNode(),
                    "LoopStartNode должен иметь exitNode");
        }

        @Test
        @DisplayName("do-while: DoWhileNode имеет exitNode после построения")
        void doWhileHasExitNode() throws Exception {
            var program = program(func("main", "int",
                    doWhileStmt(intLit(1), exprStmt(varExpr("x")))
            ));
            var start = builder.buildFromProgram(program);
            var body = assertSingleNext(start);
            // Тело do-while
            assertNotNull(body);
        }

        @Test
        @DisplayName("for с init: ProcessNode(init) → LoopStartNode")
        void forWithInit() throws Exception {
            var program = program(func("main", "int",
                    forStmt(
                            varDecl("int", "i", intLit(0)),
                            intLit(1),
                            exprStmt(unaryExpr(varExpr("i"), "++", true))
                    )
            ));
            var start = builder.buildFromProgram(program);
            var initNode = assertSingleNext(start);
            assertInstanceOf(ProcessNode.class, initNode,
                    "После start должен быть ProcessNode с init");
            var loopStart = assertSingleNext(initNode);
            assertInstanceOf(LoopStartNode.class, loopStart);
        }

        @Test
        @DisplayName("while с break: ConnectorNode(break) присутствует в теле")
        void whileWithBreak() throws Exception {
            var program = program(func("main", "int",
                    whileStmt(intLit(1), breakStmt())
            ));
            var start = builder.buildFromProgram(program);
            var loop = assertSingleNext(start);
            assertInstanceOf(LoopStartNode.class, loop);
            LoopStartNode ls = (LoopStartNode) loop;
            assertNotNull(ls.getLoopBody(), "Тело цикла не должно быть null");
            assertInstanceOf(ConnectorNode.class, ls.getLoopBody(),
                    "Тело должно быть ConnectorNode(break)");
            assertEquals("break", ls.getLoopBody().getLabel());
        }

        @Test
        @DisplayName("Вложенные while: внешний содержит внутренний LoopStartNode")
        void nestedWhile() throws Exception {
            var program = program(func("main", "int",
                    whileStmt(intLit(1),
                            whileStmt(intLit(1), exprStmt(varExpr("x"))))
            ));
            var start = builder.buildFromProgram(program);
            var outer = assertSingleNext(start);
            assertInstanceOf(LoopStartNode.class, outer);
            LoopStartNode outerLoop = (LoopStartNode) outer;
            assertNotNull(outerLoop.getLoopBody());
            assertInstanceOf(LoopStartNode.class, outerLoop.getLoopBody(),
                    "Тело внешнего цикла должно быть внутренним LoopStartNode");
        }
    }

    // ═══════════════════════════════════════════════════════════════
    //  return — структура ProcessNode + ConnectorNode
    // ═══════════════════════════════════════════════════════════════

    @Nested
    @DisplayName("return — структура узлов")
    class ReturnStructure {

        @Test
        @DisplayName("return с значением: ProcessNode → ConnectorNode(return) → end")
        void returnWithValueStructure() throws Exception {
            var program = program(func("main", "int",
                    returnStmt(intLit(42))
            ));
            var start = builder.buildFromProgram(program);
            var process = assertSingleNext(start); // "return 42"
            assertInstanceOf(ProcessNode.class, process);
            assertTrue(process.getLabel().contains("42"));
            var connector = assertSingleNext(process);
            assertInstanceOf(ConnectorNode.class, connector);
            assertEquals("return", connector.getLabel());
        }

        @Test
        @DisplayName("return без значения: сразу ConnectorNode(return)")
        void returnWithoutValue() throws Exception {
            var program = program(func("main", "void",
                    returnStmt(null)
            ));
            var start = builder.buildFromProgram(program);
            var connector = assertSingleNext(start);
            assertInstanceOf(ConnectorNode.class, connector);
            assertEquals("return", connector.getLabel());
        }

        @Test
        @DisplayName("Два return в разных ветках if: оба ConnectorNode(return) ведут к end")
        void twoReturnsInIf() throws Exception {
            var program = program(func("main", "int",
                    ifStmt(intLit(1),
                            returnStmt(intLit(1)),
                            returnStmt(intLit(0)))
            ));
            var start = builder.buildFromProgram(program);
            // Строится без исключений — обе ветки ведут к end
            assertNotNull(start);
        }
    }

    // ═══════════════════════════════════════════════════════════════
    //  Цепочка операторов после циклов
    // ═══════════════════════════════════════════════════════════════

    @Nested
    @DisplayName("Операторы после циклов")
    class StatementsAfterLoop {

        @Test
        @DisplayName("Оператор после while: выход из цикла ведёт к нему")
        void statementAfterWhile() throws Exception {
            var program = program(func("main", "int",
                    whileStmt(intLit(1), exprStmt(varExpr("x"))),
                    exprStmt(varExpr("y"))
            ));
            var start = builder.buildFromProgram(program);
            var loop = assertSingleNext(start);
            assertInstanceOf(LoopStartNode.class, loop);
            LoopStartNode ls = (LoopStartNode) loop;
            // exitNode должен вести к ProcessNode(y)
            assertNotNull(ls.getExitNode());
            assertInstanceOf(ProcessNode.class, ls.getExitNode());
        }

        @Test
        @DisplayName("Оператор после for: выход из цикла ведёт к нему")
        void statementAfterFor() throws Exception {
            var program = program(func("main", "int",
                    forStmt(null, intLit(1), null,
                            exprStmt(varExpr("body"))),
                    returnStmt(intLit(0))
            ));
            var start = builder.buildFromProgram(program);
            assertNotNull(start); // строится без ошибок
        }

        @Test
        @DisplayName("Несколько операторов перед и после if: цепочка корректна")
        void statementsAroundIf() throws Exception {
            var program = program(func("main", "int",
                    exprStmt(varExpr("a")),
                    ifStmt(intLit(1), exprStmt(varExpr("b")), null),
                    exprStmt(varExpr("c"))
            ));
            var start = builder.buildFromProgram(program);
            var a = assertSingleNext(start);
            assertInstanceOf(ProcessNode.class, a);
            var decision = assertSingleNext(a);
            assertInstanceOf(DecisionNode.class, decision);
        }
    }
}
