package com.metrics.calculator;

import com.metrics.ast.*;
import com.metrics.model.FunctionMetrics;
import com.metrics.model.ProgramMetrics;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;

import java.util.ArrayList;
import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;

/**
 * Unit-тесты MetricsCalculator.
 * Тестируем: LOC, CC, max nesting, call count, return count, goto count.
 */
@DisplayName("MetricsCalculator — юнит-тесты")
class MetricsCalculatorTest {

    private MetricsCalculator calculator;

    @BeforeEach
    void setUp() {
        calculator = new MetricsCalculator();
    }

    // ─────────────────────────────────────────────────────────────
    //  Вспомогательные фабричные методы
    // ─────────────────────────────────────────────────────────────

    private Program programWith(FunctionDecl... fns) {
        Program p = new Program();
        p.setDeclarations(List.of(fns));
        return p;
    }

    private FunctionDecl fn(String name, Statement body, int startLine, int endLine) {
        FunctionDecl fn = new FunctionDecl();
        fn.setName(name);
        fn.setBody((BlockStmt) body);
        ASTLocation loc = new ASTLocation();
        loc.setLine(startLine);
        loc.setEndLine(endLine);
        fn.setLocation(loc);
        return fn;
    }

    private FunctionDecl fnNoBody(String name) {
        FunctionDecl fn = new FunctionDecl();
        fn.setName(name);
        fn.setBody(null);
        return fn;
    }

    private FunctionDecl fnWithParams(String name, int paramCount) {
        FunctionDecl fn = new FunctionDecl();
        fn.setName(name);
        fn.setBody(block());

        List<ASTNodes> params = new ArrayList<>();
        for (int i = 0; i < paramCount; i++) {
            params.add(new ASTNodes() {}); // или конкретная реализация
        }

        fn.setParameters(params);

        ASTLocation loc = new ASTLocation();
        loc.setLine(1);
        loc.setEndLine(5);
        fn.setLocation(loc);

        return fn;
    }

    private BlockStmt block(Statement... stmts) {
        BlockStmt b = new BlockStmt();
        b.setStatements(List.of(stmts));
        return b;
    }

    private IfStmt ifStmt(Expression cond, Statement then, Statement elseBlock) {
        IfStmt s = new IfStmt();
        s.setCondition(cond);
        s.setThenBlock(then);
        s.setElseBlock(elseBlock);
        return s;
    }

    private WhileStmt whileStmt(Expression cond, Statement body) {
        WhileStmt s = new WhileStmt();
        s.setCondition(cond);
        s.setBody(body);
        return s;
    }

    private ForStmt forStmt(Expression cond, Statement body) {
        ForStmt s = new ForStmt();
        s.setCondition(cond);
        s.setBody(body);
        return s;
    }

    private DoWhileStmt doWhileStmt(Expression cond, Statement body) {
        DoWhileStmt s = new DoWhileStmt();
        s.setCondition(cond);
        s.setBody(body);
        return s;
    }

    private ReturnStmt returnStmt() {
        ReturnStmt r = new ReturnStmt();
        r.setValue(null);
        return r;
    }

    private ReturnStmt returnStmt(Expression val) {
        ReturnStmt r = new ReturnStmt();
        r.setValue(val);
        return r;
    }

    private ExprStmt exprStmt(Expression e) {
        ExprStmt s = new ExprStmt();
        s.setExpression(e);
        return s;
    }

    private BinaryExpr binExpr(Expression left, String op, Expression right) {
        BinaryExpr e = new BinaryExpr();
        e.setLeft(left);
        e.setOperator(op);
        e.setRight(right);
        return e;
    }

    private CallExpr callExpr(String name, Expression... args) {
        CallExpr e = new CallExpr();
        e.setFunctionName(name);
        e.setArguments(List.of(args));
        return e;
    }

    private VariableDecl varDecl() {
        return new VariableDecl();
    }

    private VariableDecl varDeclWithInit(Expression init) {
        VariableDecl v = new VariableDecl();
        v.setInitExpr(init);
        return v;
    }

    private GotoStmt gotoStmt() {
        return new GotoStmt();
    }

    private LabelStmt labelStmt(Statement inner) {
        LabelStmt l = new LabelStmt();
        l.setStatement(inner);
        return l;
    }

    private IntLiteral intLit(int v) {
        IntLiteral l = new IntLiteral();
        l.setValue(v);
        return l;
    }

    private VariableExpr varExpr(String name) {
        VariableExpr e = new VariableExpr();
        e.setName(name);
        return e;
    }

    // ─────────────────────────────────────────────────────────────
    //  Базовые случаи
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("Базовые случаи")
    class BasicCases {

        @Test
        @DisplayName("Пустая программа: 0 функций, 0 глобальных переменных")
        void emptyProgram() {
            Program p = new Program();
            p.setDeclarations(List.of());
            ProgramMetrics m = calculator.calculate(p);
            assertThat(m.getFunctionCount()).isZero();
            assertThat(m.getGlobalVarCount()).isZero();
            assertThat(m.getFunctions()).isEmpty();
        }

        @Test
        @DisplayName("Программа с одной функцией без тела: CC=1, LOC=1")
        void functionWithoutBody() {
            Program p = programWith(fnNoBody("stub"));
            ProgramMetrics m = calculator.calculate(p);
            assertThat(m.getFunctionCount()).isEqualTo(1);
            FunctionMetrics f = m.getFunctions().get(0);
            assertThat(f.getCyclomaticComplexity()).isEqualTo(1);
            assertThat(f.getLoc()).isEqualTo(1);
        }

        @Test
        @DisplayName("Глобальные переменные считаются отдельно от функций")
        void globalVarsCount() {
            Program p = new Program();
            p.setDeclarations(List.of(varDecl(), varDecl(), varDecl()));
            ProgramMetrics m = calculator.calculate(p);
            assertThat(m.getGlobalVarCount()).isEqualTo(3);
            assertThat(m.getFunctionCount()).isZero();
        }

        @Test
        @DisplayName("Параметры функции считаются корректно")
        void parameterCount() {
            FunctionDecl f = fnWithParams("foo", 3);
            Program p = programWith(f);
            FunctionMetrics m = calculator.calculate(p).getFunctions().get(0);
            assertThat(m.getParameterCount()).isEqualTo(3);
        }

        @Test
        @DisplayName("Функция без параметров: parameterCount = 0")
        void noParameters() {
            FunctionDecl f = fn("bar", block(), 1, 5);
            f.setParameters(null);
            Program p = programWith(f);
            FunctionMetrics m = calculator.calculate(p).getFunctions().get(0);
            assertThat(m.getParameterCount()).isZero();
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  LOC
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("LOC (строки кода)")
    class LocCalculation {

        @Test
        @DisplayName("LOC = endLine - startLine + 1")
        void locFromLocation() {
            FunctionDecl f = fn("foo", block(), 5, 15);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getLoc()).isEqualTo(11); // 15 - 5 + 1
        }

        @Test
        @DisplayName("LOC минимум 1, если location == null")
        void locNoLocation() {
            FunctionDecl f = new FunctionDecl();
            f.setName("foo");
            f.setBody(block());
            f.setLocation(null);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getLoc()).isEqualTo(1);
        }

        @Test
        @DisplayName("LOC минимум 1, даже если startLine > endLine")
        void locMinimumOne() {
            FunctionDecl f = fn("foo", block(), 10, 8); // endLine < startLine
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getLoc()).isEqualTo(1);
        }

        @Test
        @DisplayName("LOC однострочной функции = 1")
        void locSingleLine() {
            FunctionDecl f = fn("tiny", block(), 7, 7);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getLoc()).isEqualTo(1);
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  Цикломатическая сложность
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("Цикломатическая сложность (CC)")
    class CyclomaticComplexity {

        @Test
        @DisplayName("Пустая функция: CC = 1")
        void emptyFunction() {
            FunctionDecl f = fn("main", block(), 1, 5);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getCyclomaticComplexity()).isEqualTo(1);
        }

        @Test
        @DisplayName("Один if: CC = 2")
        void singleIf() {
            FunctionDecl f = fn("foo", block(
                    ifStmt(intLit(1), block(), null)
            ), 1, 5);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getCyclomaticComplexity()).isEqualTo(2);
        }

        @Test
        @DisplayName("if-else: CC = 2 (else не добавляет)")
        void ifElse() {
            FunctionDecl f = fn("foo", block(
                    ifStmt(intLit(1), block(), block())
            ), 1, 10);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getCyclomaticComplexity()).isEqualTo(2);
        }

        @Test
        @DisplayName("while: CC = 2")
        void whileLoop() {
            FunctionDecl f = fn("foo", block(
                    whileStmt(intLit(1), block())
            ), 1, 5);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getCyclomaticComplexity()).isEqualTo(2);
        }

        @Test
        @DisplayName("for: CC = 2")
        void forLoop() {
            FunctionDecl f = fn("foo", block(
                    forStmt(intLit(1), block())
            ), 1, 5);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getCyclomaticComplexity()).isEqualTo(2);
        }

        @Test
        @DisplayName("do-while: CC = 2")
        void doWhileLoop() {
            FunctionDecl f = fn("foo", block(
                    doWhileStmt(intLit(1), block())
            ), 1, 5);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getCyclomaticComplexity()).isEqualTo(2);
        }

        @Test
        @DisplayName("Оператор && в условии: +1 к CC")
        void andOperatorInCondition() {
            BinaryExpr cond = binExpr(intLit(1), "&&", intLit(1));
            FunctionDecl f = fn("foo", block(
                    ifStmt(cond, block(), null)
            ), 1, 5);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            // 1 (base) + 1 (if) + 1 (&&) = 3
            assertThat(m.getCyclomaticComplexity()).isEqualTo(3);
        }

        @Test
        @DisplayName("Оператор || в условии: +1 к CC")
        void orOperatorInCondition() {
            BinaryExpr cond = binExpr(intLit(0), "||", intLit(1));
            FunctionDecl f = fn("foo", block(
                    ifStmt(cond, block(), null)
            ), 1, 5);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getCyclomaticComplexity()).isEqualTo(3);
        }

        @Test
        @DisplayName("Вложенный if: CC суммируется")
        void nestedIfs() {
            FunctionDecl f = fn("foo", block(
                    ifStmt(intLit(1),
                            block(ifStmt(intLit(2), block(), null)),
                            null)
            ), 1, 10);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            // 1 + 1 + 1 = 3
            assertThat(m.getCyclomaticComplexity()).isEqualTo(3);
        }

        @Test
        @DisplayName("if + while + for: CC = 4")
        void combinedStructures() {
            FunctionDecl f = fn("foo", block(
                    ifStmt(intLit(1), block(), null),
                    whileStmt(intLit(1), block()),
                    forStmt(intLit(1), block())
            ), 1, 20);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getCyclomaticComplexity()).isEqualTo(4);
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  Максимальная глубина вложенности
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("Максимальная глубина вложенности")
    class MaxNestingDepth {

        @Test
        @DisplayName("Пустая функция: depth = 0")
        void emptyFunction() {
            FunctionDecl f = fn("foo", block(), 1, 5);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getMaxNestingDepth()).isZero();
        }

        @Test
        @DisplayName("Один if: depth = 1")
        void singleIf() {
            FunctionDecl f = fn("foo", block(
                    ifStmt(intLit(1), block(), null)
            ), 1, 5);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getMaxNestingDepth()).isEqualTo(1);
        }

        @Test
        @DisplayName("if внутри while: depth = 2")
        void ifInsideWhile() {
            FunctionDecl f = fn("foo", block(
                    whileStmt(intLit(1),
                            block(ifStmt(intLit(1), block(), null)))
            ), 1, 10);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getMaxNestingDepth()).isEqualTo(2);
        }

        @Test
        @DisplayName("Три уровня вложенности: for → while → if, depth = 3")
        void threeNestingLevels() {
            FunctionDecl f = fn("foo", block(
                    forStmt(intLit(1),
                            block(whileStmt(intLit(1),
                                    block(ifStmt(intLit(1), block(), null)))))
            ), 1, 15);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getMaxNestingDepth()).isEqualTo(3);
        }

        @Test
        @DisplayName("Два параллельных if: depth = 1 (максимум, не сумма)")
        void parallelIfs() {
            FunctionDecl f = fn("foo", block(
                    ifStmt(intLit(1), block(), null),
                    ifStmt(intLit(2), block(), null)
            ), 1, 10);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getMaxNestingDepth()).isEqualTo(1);
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  Подсчёт вызовов функций
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("Количество вызовов функций (callCount)")
    class CallCount {

        @Test
        @DisplayName("Нет вызовов: callCount = 0")
        void noCalls() {
            FunctionDecl f = fn("foo", block(), 1, 5);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getCallCount()).isZero();
        }

        @Test
        @DisplayName("Один вызов в ExprStmt: callCount = 1")
        void singleCallInExprStmt() {
            FunctionDecl f = fn("foo", block(
                    exprStmt(callExpr("bar"))
            ), 1, 5);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getCallCount()).isEqualTo(1);
        }

        @Test
        @DisplayName("Два вызова: callCount = 2")
        void twoCalls() {
            FunctionDecl f = fn("foo", block(
                    exprStmt(callExpr("a")),
                    exprStmt(callExpr("b"))
            ), 1, 5);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getCallCount()).isEqualTo(2);
        }

        @Test
        @DisplayName("Вызов в условии while считается")
        void callInWhileCondition() {
            FunctionDecl f = fn("foo", block(
                    whileStmt(callExpr("isReady"), block())
            ), 1, 5);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getCallCount()).isEqualTo(1);
        }

        @Test
        @DisplayName("Вызов в return считается")
        void callInReturn() {
            FunctionDecl f = fn("foo", block(
                    returnStmt(callExpr("compute"))
            ), 1, 5);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getCallCount()).isEqualTo(1);
        }

        @Test
        @DisplayName("Вызов в инициализации переменной считается")
        void callInVarDecl() {
            FunctionDecl f = fn("foo", block(
                    varDeclWithInit(callExpr("getValue"))
            ), 1, 5);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getCallCount()).isEqualTo(1);
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  Подсчёт return
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("Количество return (returnCount)")
    class ReturnCount {

        @Test
        @DisplayName("Нет return: returnCount = 0")
        void noReturns() {
            FunctionDecl f = fn("foo", block(), 1, 5);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getReturnCount()).isZero();
        }

        @Test
        @DisplayName("Один return: returnCount = 1")
        void oneReturn() {
            FunctionDecl f = fn("foo", block(returnStmt()), 1, 5);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getReturnCount()).isEqualTo(1);
        }

        @Test
        @DisplayName("return в обеих ветках if: returnCount = 2")
        void returnInBothBranches() {
            FunctionDecl f = fn("foo", block(
                    ifStmt(intLit(1), block(returnStmt(intLit(1))), block(returnStmt(intLit(0))))
            ), 1, 10);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getReturnCount()).isEqualTo(2);
        }

        @Test
        @DisplayName("return внутри while: returnCount = 1")
        void returnInsideWhile() {
            FunctionDecl f = fn("foo", block(
                    whileStmt(intLit(1), block(returnStmt()))
            ), 1, 5);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getReturnCount()).isEqualTo(1);
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  Подсчёт goto
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("Количество goto (gotoCount)")
    class GotoCount {

        @Test
        @DisplayName("Нет goto: gotoCount = 0")
        void noGotos() {
            FunctionDecl f = fn("foo", block(), 1, 5);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getGotoCount()).isZero();
        }

        @Test
        @DisplayName("Один goto: gotoCount = 1")
        void oneGoto() {
            FunctionDecl f = fn("foo", block(gotoStmt()), 1, 5);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getGotoCount()).isEqualTo(1);
        }

        @Test
        @DisplayName("goto внутри if: gotoCount = 1")
        void gotoInsideIf() {
            FunctionDecl f = fn("foo", block(
                    ifStmt(intLit(1), block(gotoStmt()), null)
            ), 1, 5);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getGotoCount()).isEqualTo(1);
        }

        @Test
        @DisplayName("goto внутри LabelStmt: gotoCount = 1")
        void gotoInsideLabelStmt() {
            FunctionDecl f = fn("foo", block(
                    labelStmt(gotoStmt())
            ), 1, 5);
            FunctionMetrics m = calculator.calculate(programWith(f)).getFunctions().get(0);
            assertThat(m.getGotoCount()).isEqualTo(1);
        }
    }

    // ─────────────────────────────────────────────────────────────
    //  Несколько функций в программе
    // ─────────────────────────────────────────────────────────────

    @Nested
    @DisplayName("Несколько функций")
    class MultipleFunctions {

        @Test
        @DisplayName("Программа с 2 функциями: метрики независимы")
        void twoFunctions() {
            FunctionDecl f1 = fn("main", block(ifStmt(intLit(1), block(), null)), 1, 5);
            FunctionDecl f2 = fn("helper", block(returnStmt()), 6, 8);
            ProgramMetrics m = calculator.calculate(programWith(f1, f2));

            assertThat(m.getFunctionCount()).isEqualTo(2);
            FunctionMetrics m1 = m.getFunctions().stream()
                    .filter(x -> "main".equals(x.getFunctionName())).findFirst().orElseThrow();
            FunctionMetrics m2 = m.getFunctions().stream()
                    .filter(x -> "helper".equals(x.getFunctionName())).findFirst().orElseThrow();

            assertThat(m1.getCyclomaticComplexity()).isEqualTo(2);
            assertThat(m2.getReturnCount()).isEqualTo(1);
        }
    }
}