package com.metrics.calculator;

import com.metrics.ast.*;
import com.metrics.model.FunctionMetrics;
import com.metrics.model.ProgramMetrics;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;

/**
 * Дополнительные тесты на граничные случаи, нацеленные на выживших мутантов.
 *
 * Покрывают:
 *  - UnaryExpr, AssignmentExpr, ArrayAccessExpr в CC и callCount
 *  - DoWhileStmt в returnCount и gotoCount
 *  - LabelStmt во всех счётчиках
 *  - Граничные значения LOC (end == start, end == start-1)
 *  - Цепочки && и || в CC
 *  - Вызовы внутри аргументов CallExpr
 *  - AssignmentExpr с CallExpr в left и right
 *  - ArrayAccessExpr в countCallsExpr
 *  - ForStmt без condition в CC
 *  - IfStmt с else в nesting depth
 */
@DisplayName("MetricsCalculator — мутационные граничные тесты")
class MetricsCalculatorMutationTest {

    private MetricsCalculator calculator;

    @BeforeEach
    void setUp() {
        calculator = new MetricsCalculator();
    }

    // ── Фабричные методы ──────────────────────────────────────────

    private Program prog(FunctionDecl... fns) {
        Program p = new Program();
        p.setDeclarations(List.of(fns));
        return p;
    }

    private FunctionDecl fn(Statement body) {
        FunctionDecl f = new FunctionDecl();
        f.setName("f");
        f.setBody((BlockStmt) body);
        ASTLocation loc = new ASTLocation();
        loc.setLine(1); loc.setEndLine(10);
        f.setLocation(loc);
        return f;
    }

    private FunctionDecl fnAt(Statement body, int start, int end) {
        FunctionDecl f = new FunctionDecl();
        f.setName("f");
        f.setBody((BlockStmt) body);
        ASTLocation loc = new ASTLocation();
        loc.setLine(start); loc.setEndLine(end);
        f.setLocation(loc);
        return f;
    }

    private BlockStmt block(Statement... stmts) {
        BlockStmt b = new BlockStmt();
        b.setStatements(List.of(stmts));
        return b;
    }

    private IfStmt ifS(Expression cond, Statement then, Statement el) {
        IfStmt s = new IfStmt(); s.setCondition(cond);
        s.setThenBlock(then); s.setElseBlock(el); return s;
    }

    private WhileStmt whileS(Expression cond, Statement body) {
        WhileStmt s = new WhileStmt(); s.setCondition(cond); s.setBody(body); return s;
    }

    private ForStmt forS(Expression cond, Statement body) {
        ForStmt s = new ForStmt(); s.setCondition(cond); s.setBody(body); return s;
    }

    private DoWhileStmt doWhileS(Expression cond, Statement body) {
        DoWhileStmt s = new DoWhileStmt(); s.setCondition(cond); s.setBody(body); return s;
    }

    private ReturnStmt ret(Expression val) {
        ReturnStmt r = new ReturnStmt(); r.setValue(val); return r;
    }

    private ReturnStmt ret() { return ret(null); }

    private ExprStmt exprS(Expression e) {
        ExprStmt s = new ExprStmt(); s.setExpression(e); return s;
    }

    private LabelStmt labelS(Statement inner) {
        LabelStmt l = new LabelStmt(); l.setStatement(inner); return l;
    }

    private GotoStmt gotoS() { return new GotoStmt(); }

    private BinaryExpr bin(Expression l, String op, Expression r) {
        BinaryExpr e = new BinaryExpr();
        e.setLeft(l); e.setOperator(op); e.setRight(r); return e;
    }

    private UnaryExpr unary(Expression operand) {
        UnaryExpr e = new UnaryExpr(); e.setOperand(operand); return e;
    }

    private AssignmentExpr assign(Expression left, Expression right) {
        AssignmentExpr e = new AssignmentExpr();
        e.setLeft(left); e.setRight(right); return e;
    }

    private CallExpr call(String name, Expression... args) {
        CallExpr e = new CallExpr();
        e.setFunctionName(name); e.setArguments(List.of(args)); return e;
    }

    private ArrayAccessExpr arrayAccess(Expression array, Expression index) {
        ArrayAccessExpr e = new ArrayAccessExpr();
        e.setArray(array); e.setIndex(index); return e;
    }

    private IntLiteral lit(int v) {
        IntLiteral l = new IntLiteral(); l.setValue(v); return l;
    }

    private VariableExpr var(String n) {
        VariableExpr e = new VariableExpr(); e.setName(n); return e;
    }

    private FunctionMetrics calc(Statement body) {
        return calculator.calculate(prog(fn(body))).getFunctions().get(0);
    }

    // ═══════════════════════════════════════════════════════════════
    //  LOC — граничные значения
    // ═══════════════════════════════════════════════════════════════

    @Nested
    @DisplayName("LOC граничные значения")
    class LocBoundary {

        @Test
        @DisplayName("LOC = 2 когда end = start + 1")
        void locTwoLines() {
            FunctionMetrics m = calculator.calculate(prog(fnAt(block(), 5, 6))).getFunctions().get(0);
            assertThat(m.getLoc()).isEqualTo(2);
        }

        @Test
        @DisplayName("LOC = 1 когда end == start (не ноль)")
        void locExactlyOne() {
            FunctionMetrics m = calculator.calculate(prog(fnAt(block(), 3, 3))).getFunctions().get(0);
            assertThat(m.getLoc()).isEqualTo(1);
        }

        @Test
        @DisplayName("LOC = 1 когда end = start - 1 (Math.max защита)")
        void locEndBeforeStart() {
            FunctionMetrics m = calculator.calculate(prog(fnAt(block(), 10, 9))).getFunctions().get(0);
            assertThat(m.getLoc()).isEqualTo(1); // Math.max(1, 9-10+1) = Math.max(1,0) = 1
        }

        @Test
        @DisplayName("LOC корректен при большом диапазоне строк")
        void locLarge() {
            FunctionMetrics m = calculator.calculate(prog(fnAt(block(), 1, 100))).getFunctions().get(0);
            assertThat(m.getLoc()).isEqualTo(100);
        }
    }

    // ═══════════════════════════════════════════════════════════════
    //  CC — UnaryExpr, AssignmentExpr, ArrayAccessExpr в условиях
    // ═══════════════════════════════════════════════════════════════

    @Nested
    @DisplayName("CC — выражения в условиях")
    class CcExpressions {

        @Test
        @DisplayName("UnaryExpr внутри условия if не добавляет CC")
        void unaryInIfCondition() {
            // if (!x) — унарный не добавляет decision point
            FunctionMetrics m = calc(block(ifS(unary(var("x")), block(), null)));
            assertThat(m.getCyclomaticComplexity()).isEqualTo(2); // 1 base + 1 if
        }

        @Test
        @DisplayName("UnaryExpr содержащий && добавляет CC")
        void unaryWrappingAndExpr() {
            // Унарный оборачивает && — && внутри всё равно считается
            BinaryExpr andExpr = bin(lit(1), "&&", lit(0));
            UnaryExpr unaryAnd = unary(andExpr);
            FunctionMetrics m = calc(block(ifS(unaryAnd, block(), null)));
            // 1 base + 1 if + 1 &&
            assertThat(m.getCyclomaticComplexity()).isEqualTo(3);
        }

        @Test
        @DisplayName("AssignmentExpr в ExprStmt с && в right не добавляет CC")
        void assignmentWithNoBooleanOps() {
            // x = 5 — нет булевых операторов
            AssignmentExpr a = assign(var("x"), lit(5));
            FunctionMetrics m = calc(block(exprS(a)));
            assertThat(m.getCyclomaticComplexity()).isEqualTo(1);
        }

        @Test
        @DisplayName("AssignmentExpr с && в правой части добавляет CC")
        void assignmentWithAndInRight() {
            // x = (a && b) — правая часть содержит &&
            AssignmentExpr a = assign(var("x"), bin(var("a"), "&&", var("b")));
            FunctionMetrics m = calc(block(exprS(a)));
            // 1 base + 1 &&
            assertThat(m.getCyclomaticComplexity()).isEqualTo(2);
        }

        @Test
        @DisplayName("ArrayAccessExpr с && в индексе добавляет CC")
        void arrayAccessWithAndInIndex() {
            // arr[a && b] — индекс содержит &&
            ArrayAccessExpr arr = arrayAccess(var("arr"), bin(var("a"), "&&", var("b")));
            FunctionMetrics m = calc(block(ifS(arr, block(), null)));
            // 1 base + 1 if + 1 &&
            assertThat(m.getCyclomaticComplexity()).isEqualTo(3);
        }

        @Test
        @DisplayName("Цепочка && && добавляет 2 к CC")
        void doubleAnd() {
            // (a && b && c) — два &&
            BinaryExpr inner = bin(var("a"), "&&", var("b"));
            BinaryExpr outer = bin(inner, "&&", var("c"));
            FunctionMetrics m = calc(block(ifS(outer, block(), null)));
            // 1 + 1(if) + 2(&&) = 4
            assertThat(m.getCyclomaticComplexity()).isEqualTo(4);
        }

        @Test
        @DisplayName("Смешанный && и ||: оба добавляют CC")
        void mixedAndOr() {
            BinaryExpr andPart = bin(var("a"), "&&", var("b"));
            BinaryExpr orPart = bin(andPart, "||", var("c"));
            FunctionMetrics m = calc(block(ifS(orPart, block(), null)));
            // 1 + 1(if) + 1(&&) + 1(||) = 4
            assertThat(m.getCyclomaticComplexity()).isEqualTo(4);
        }

        @Test
        @DisplayName("ForStmt без condition: +1 к CC (только за структуру)")
        void forWithoutCondition() {
            ForStmt f = new ForStmt();
            f.setCondition(null);
            f.setBody(block());
            FunctionMetrics m = calc(block(f));
            assertThat(m.getCyclomaticComplexity()).isEqualTo(2); // 1 + 1(for)
        }

        @Test
        @DisplayName("ReturnStmt с && в значении добавляет CC")
        void returnWithAndExpr() {
            ReturnStmt r = ret(bin(var("a"), "&&", var("b")));
            FunctionMetrics m = calc(block(r));
            assertThat(m.getCyclomaticComplexity()).isEqualTo(2); // 1 + 1(&&)
        }

        @Test
        @DisplayName("LabelStmt оборачивающий if: CC считается из внутреннего if")
        void labelWrappingIf() {
            FunctionMetrics m = calc(block(labelS(ifS(lit(1), block(), null))));
            assertThat(m.getCyclomaticComplexity()).isEqualTo(2); // 1 + 1(if)
        }

        @Test
        @DisplayName("CallExpr с && в аргументе добавляет CC")
        void callExprWithAndInArg() {
            CallExpr c = call("foo", bin(var("a"), "&&", var("b")));
            FunctionMetrics m = calc(block(exprS(c)));
            assertThat(m.getCyclomaticComplexity()).isEqualTo(2); // 1 + 1(&&)
        }
    }

    // ═══════════════════════════════════════════════════════════════
    //  Max nesting — DoWhile, else ветка, LabelStmt
    // ═══════════════════════════════════════════════════════════════

    @Nested
    @DisplayName("Max nesting — дополнительные случаи")
    class MaxNestingExtra {

        @Test
        @DisplayName("DoWhile: depth = 1")
        void doWhileDepth() {
            FunctionMetrics m = calc(block(doWhileS(lit(1), block())));
            assertThat(m.getMaxNestingDepth()).isEqualTo(1);
        }

        @Test
        @DisplayName("DoWhile с вложенным if: depth = 2")
        void doWhileWithIf() {
            FunctionMetrics m = calc(block(
                    doWhileS(lit(1), block(ifS(lit(1), block(), null)))
            ));
            assertThat(m.getMaxNestingDepth()).isEqualTo(2);
        }

        @Test
        @DisplayName("if-else: depth по ветке else тоже = 1")
        void ifElseBothDepthOne() {
            FunctionMetrics m = calc(block(
                    ifS(lit(1), block(), block())
            ));
            assertThat(m.getMaxNestingDepth()).isEqualTo(1);
        }

        @Test
        @DisplayName("Глубокий else vs мелкий then: берём максимум из else")
        void deepElseBranch() {
            // then: depth 1, else: if → if = depth 3
            FunctionMetrics m = calc(block(
                    ifS(lit(1),
                        block(),
                        block(ifS(lit(2),
                            block(ifS(lit(3), block(), null)),
                            null))
                    )
            ));
            assertThat(m.getMaxNestingDepth()).isEqualTo(3);
        }

        @Test
        @DisplayName("LabelStmt не увеличивает depth")
        void labelDoesNotIncreaseDepth() {
            FunctionMetrics m = calc(block(labelS(ret())));
            assertThat(m.getMaxNestingDepth()).isEqualTo(0);
        }

        @Test
        @DisplayName("LabelStmt оборачивающий if: depth = 1")
        void labelWrappingIfDepth() {
            FunctionMetrics m = calc(block(labelS(ifS(lit(1), block(), null))));
            assertThat(m.getMaxNestingDepth()).isEqualTo(1);
        }

        @Test
        @DisplayName("for внутри if: depth = 2")
        void forInsideIf() {
            FunctionMetrics m = calc(block(
                    ifS(lit(1), block(forS(lit(1), block())), null)
            ));
            assertThat(m.getMaxNestingDepth()).isEqualTo(2);
        }
    }

    // ═══════════════════════════════════════════════════════════════
    //  callCount — AssignmentExpr, ArrayAccessExpr, UnaryExpr, DoWhile
    // ═══════════════════════════════════════════════════════════════

    @Nested
    @DisplayName("callCount — дополнительные случаи")
    class CallCountExtra {

        @Test
        @DisplayName("CallExpr с вызовом в аргументе: callCount = 2")
        void nestedCallInArgs() {
            // foo(bar()) — два вызова
            CallExpr inner = call("bar");
            CallExpr outer = call("foo", inner);
            FunctionMetrics m = calc(block(exprS(outer)));
            assertThat(m.getCallCount()).isEqualTo(2);
        }

        @Test
        @DisplayName("AssignmentExpr с вызовом в left: callCount = 1")
        void callInAssignLeft() {
            // Вызов в левой части присваивания (напр. arr[foo()] = x)
            AssignmentExpr a = assign(call("foo"), var("x"));
            FunctionMetrics m = calc(block(exprS(a)));
            assertThat(m.getCallCount()).isEqualTo(1);
        }

        @Test
        @DisplayName("AssignmentExpr с вызовом в right: callCount = 1")
        void callInAssignRight() {
            AssignmentExpr a = assign(var("x"), call("compute"));
            FunctionMetrics m = calc(block(exprS(a)));
            assertThat(m.getCallCount()).isEqualTo(1);
        }

        @Test
        @DisplayName("AssignmentExpr с вызовами и в left и в right: callCount = 2")
        void callInBothSidesOfAssign() {
            AssignmentExpr a = assign(call("getRef"), call("compute"));
            FunctionMetrics m = calc(block(exprS(a)));
            assertThat(m.getCallCount()).isEqualTo(2);
        }

        @Test
        @DisplayName("ArrayAccessExpr с вызовом в array: callCount = 1")
        void callInArrayExpr() {
            ArrayAccessExpr arr = arrayAccess(call("getArr"), lit(0));
            FunctionMetrics m = calc(block(exprS(arr)));
            assertThat(m.getCallCount()).isEqualTo(1);
        }

        @Test
        @DisplayName("ArrayAccessExpr с вызовом в index: callCount = 1")
        void callInArrayIndex() {
            ArrayAccessExpr arr = arrayAccess(var("arr"), call("getIdx"));
            FunctionMetrics m = calc(block(exprS(arr)));
            assertThat(m.getCallCount()).isEqualTo(1);
        }

        @Test
        @DisplayName("UnaryExpr с вызовом внутри: callCount = 1")
        void callInUnary() {
            UnaryExpr u = unary(call("check"));
            FunctionMetrics m = calc(block(exprS(u)));
            assertThat(m.getCallCount()).isEqualTo(1);
        }

        @Test
        @DisplayName("Вызов в условии DoWhile считается")
        void callInDoWhileCondition() {
            FunctionMetrics m = calc(block(
                    doWhileS(call("hasNext"), block())
            ));
            assertThat(m.getCallCount()).isEqualTo(1);
        }

        @Test
        @DisplayName("Вызов в условии if считается")
        void callInIfCondition() {
            FunctionMetrics m = calc(block(
                    ifS(call("check"), block(), null)
            ));
            assertThat(m.getCallCount()).isEqualTo(1);
        }

        @Test
        @DisplayName("Вызов в условии for считается")
        void callInForCondition() {
            FunctionMetrics m = calc(block(
                    forS(call("hasMore"), block())
            ));
            assertThat(m.getCallCount()).isEqualTo(1);
        }

        @Test
        @DisplayName("Вызов в BinaryExpr (a + foo()): callCount = 1")
        void callInBinaryExpr() {
            BinaryExpr b = bin(var("a"), "+", call("foo"));
            FunctionMetrics m = calc(block(exprS(b)));
            assertThat(m.getCallCount()).isEqualTo(1);
        }

        @Test
        @DisplayName("Вызов в VariableDecl initExpr: callCount = 1")
        void callInVarDeclInit() {
            VariableDecl v = new VariableDecl();
            v.setInitExpr(call("init"));
            FunctionMetrics m = calc(block(v));
            assertThat(m.getCallCount()).isEqualTo(1);
        }

        @Test
        @DisplayName("LabelStmt с вызовом внутри: callCount = 1")
        void callInLabelStmt() {
            FunctionMetrics m = calc(block(labelS(exprS(call("foo")))));
            assertThat(m.getCallCount()).isEqualTo(1);
        }
    }

    // ═══════════════════════════════════════════════════════════════
    //  returnCount — DoWhile, ForStmt, LabelStmt
    // ═══════════════════════════════════════════════════════════════

    @Nested
    @DisplayName("returnCount — дополнительные случаи")
    class ReturnCountExtra {

        @Test
        @DisplayName("return внутри for: returnCount = 1")
        void returnInsideFor() {
            FunctionMetrics m = calc(block(forS(lit(1), block(ret()))));
            assertThat(m.getReturnCount()).isEqualTo(1);
        }

        @Test
        @DisplayName("return внутри do-while: returnCount = 1")
        void returnInsideDoWhile() {
            FunctionMetrics m = calc(block(doWhileS(lit(1), block(ret()))));
            assertThat(m.getReturnCount()).isEqualTo(1);
        }

        @Test
        @DisplayName("return внутри LabelStmt: returnCount = 1")
        void returnInsideLabel() {
            FunctionMetrics m = calc(block(labelS(ret())));
            assertThat(m.getReturnCount()).isEqualTo(1);
        }

        @Test
        @DisplayName("return в then без else: returnCount = 1")
        void returnOnlyInThen() {
            FunctionMetrics m = calc(block(
                    ifS(lit(1), block(ret(lit(1))), null)
            ));
            assertThat(m.getReturnCount()).isEqualTo(1);
        }

        @Test
        @DisplayName("Три return в разных ветках: returnCount = 3")
        void threeReturns() {
            FunctionMetrics m = calc(block(
                    ifS(lit(1), block(ret(lit(1))), block(ret(lit(0)))),
                    ret(lit(2))
            ));
            assertThat(m.getReturnCount()).isEqualTo(3);
        }

        @Test
        @DisplayName("Вложенные блоки с return суммируются")
        void nestedBlocksReturn() {
            FunctionMetrics m = calc(block(
                    whileS(lit(1), block(
                            forS(lit(1), block(ret()))
                    ))
            ));
            assertThat(m.getReturnCount()).isEqualTo(1);
        }
    }

    // ═══════════════════════════════════════════════════════════════
    //  gotoCount — DoWhile, ForStmt, LabelStmt, IfStmt else
    // ═══════════════════════════════════════════════════════════════

    @Nested
    @DisplayName("gotoCount — дополнительные случаи")
    class GotoCountExtra {

        @Test
        @DisplayName("goto внутри for: gotoCount = 1")
        void gotoInsideFor() {
            FunctionMetrics m = calc(block(forS(lit(1), block(gotoS()))));
            assertThat(m.getGotoCount()).isEqualTo(1);
        }

        @Test
        @DisplayName("goto внутри do-while: gotoCount = 1")
        void gotoInsideDoWhile() {
            FunctionMetrics m = calc(block(doWhileS(lit(1), block(gotoS()))));
            assertThat(m.getGotoCount()).isEqualTo(1);
        }

        @Test
        @DisplayName("goto в else ветке if: gotoCount = 1")
        void gotoInElseBranch() {
            FunctionMetrics m = calc(block(
                    ifS(lit(1), block(), block(gotoS()))
            ));
            assertThat(m.getGotoCount()).isEqualTo(1);
        }

        @Test
        @DisplayName("goto и в then и в else: gotoCount = 2")
        void gotoInBothBranches() {
            FunctionMetrics m = calc(block(
                    ifS(lit(1), block(gotoS()), block(gotoS()))
            ));
            assertThat(m.getGotoCount()).isEqualTo(2);
        }

        @Test
        @DisplayName("goto внутри while: gotoCount = 1")
        void gotoInsideWhile() {
            FunctionMetrics m = calc(block(whileS(lit(1), block(gotoS()))));
            assertThat(m.getGotoCount()).isEqualTo(1);
        }

        @Test
        @DisplayName("Два goto в разных операторах: gotoCount = 2")
        void twoGotos() {
            FunctionMetrics m = calc(block(gotoS(), gotoS()));
            assertThat(m.getGotoCount()).isEqualTo(2);
        }
    }

    // ═══════════════════════════════════════════════════════════════
    //  Программа: globalVarCount + functions вместе
    // ═══════════════════════════════════════════════════════════════

    @Nested
    @DisplayName("Program — globalVarCount точный подсчёт")
    class ProgramGlobalVars {

        @Test
        @DisplayName("Смесь функций и переменных: считаются корректно")
        void mixedDeclarations() {
            Program p = new Program();
            VariableDecl v1 = new VariableDecl();
            VariableDecl v2 = new VariableDecl();
            FunctionDecl f = fn(block());
            p.setDeclarations(List.of(v1, f, v2));
            ProgramMetrics m = calculator.calculate(p);
            assertThat(m.getGlobalVarCount()).isEqualTo(2);
            assertThat(m.getFunctionCount()).isEqualTo(1);
        }

        @Test
        @DisplayName("Только переменные, нет функций: функций = 0")
        void onlyVars() {
            Program p = new Program();
            p.setDeclarations(List.of(new VariableDecl(), new VariableDecl()));
            ProgramMetrics m = calculator.calculate(p);
            assertThat(m.getFunctionCount()).isZero();
            assertThat(m.getGlobalVarCount()).isEqualTo(2);
        }

        @Test
        @DisplayName("Имя функции сохраняется в метриках")
        void functionNamePreserved() {
            FunctionDecl f = fn(block());
            f.setName("myFunc");
            Program p = new Program();
            p.setDeclarations(List.of(f));
            FunctionMetrics m = calculator.calculate(p).getFunctions().get(0);
            assertThat(m.getFunctionName()).isEqualTo("myFunc");
        }
    }
}
