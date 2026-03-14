package flowchart;

import com.fasterxml.jackson.databind.ObjectMapper;
import flowchart.ast.*;

import java.util.List;

/**
 * Вспомогательный класс для создания AST в тестах.
 * Использует ObjectMapper, так как поля AST-классов приватны и сеттеров нет.
 */
public class AstBuilder {

    private static final ObjectMapper mapper = new ObjectMapper();

    // ─── Программа и функции ───────────────────────────────────────

    public static Program program(FunctionDecl... funcs) throws Exception {
        StringBuilder decls = new StringBuilder("[");
        for (int i = 0; i < funcs.length; i++) {
            if (i > 0) decls.append(",");
            decls.append(funcJson(funcs[i]));
        }
        decls.append("]");
        String json = "{\"type\":\"Program\",\"declarations\":" + decls
                + ",\"location\":" + loc(1,1,99,1) + "}";
        return mapper.readValue(json, Program.class);
    }

    public static FunctionDecl func(String name, String returnType, Statement... stmts) throws Exception {
        return func(name, returnType, List.of(), stmts);
    }

    public static FunctionDecl func(String name, String returnType, List<String[]> params, Statement... stmts) throws Exception {
        StringBuilder paramsJson = new StringBuilder("[");
        for (int i = 0; i < params.size(); i++) {
            if (i > 0) paramsJson.append(",");
            String[] p = params.get(i);
            paramsJson.append("{\"type\":{\"baseType\":\"").append(p[0])
                    .append("\",\"pointerLevel\":0,\"arraySizes\":[]},\"name\":\"").append(p[1])
                    .append("\",\"location\":").append(loc(1,1,1,10)).append("}");
        }
        paramsJson.append("]");

        StringBuilder stmtsJson = buildStmtsJson(stmts);
        String json = "{\"type\":\"FunctionDecl\",\"name\":\"" + name + "\","
                + "\"returnType\":{\"baseType\":\"" + returnType + "\",\"pointerLevel\":0,\"arraySizes\":[]},"
                + "\"parameters\":" + paramsJson + ","
                + "\"body\":{\"type\":\"BlockStmt\",\"statements\":" + stmtsJson + ",\"location\":" + loc(1,1,99,1) + "},"
                + "\"location\":" + loc(1,1,99,1) + "}";
        return mapper.readValue(json, FunctionDecl.class);
    }

    // ─── Statements ───────────────────────────────────────────────

    public static Statement varDecl(String type, String name, String initExprJson) throws Exception {
        String initPart = (initExprJson != null) ? ",\"initExpr\":" + initExprJson : "";
        String json = "{\"type\":\"VariableDecl\","
                + "\"varType\":{\"baseType\":\"" + type + "\",\"pointerLevel\":0,\"arraySizes\":[]},"
                + "\"name\":\"" + name + "\""
                + initPart + ",\"location\":" + loc(1,1,1,20) + "}";
        return mapper.readValue(json, VariableDecl.class);
    }

    public static Statement returnStmt(String valueExprJson) throws Exception {
        String valPart = (valueExprJson != null) ? ",\"value\":" + valueExprJson : "";
        String json = "{\"type\":\"ReturnStmt\"" + valPart + ",\"location\":" + loc(1,1,1,20) + "}";
        return mapper.readValue(json, ReturnStmt.class);
    }

    public static Statement exprStmt(String exprJson) throws Exception {
        String json = "{\"type\":\"ExprStmt\",\"expression\":" + exprJson
                + ",\"location\":" + loc(1,1,1,20) + "}";
        return mapper.readValue(json, ExprStmt.class);
    }

    public static Statement ifStmt(String condJson, Statement thenStmt, Statement elseStmt) throws Exception {
        String elsePart = (elseStmt != null) ? ",\"elseBlock\":" + stmtJson(elseStmt) : "";
        String json = "{\"type\":\"IfStmt\",\"condition\":" + condJson
                + ",\"thenBlock\":" + stmtJson(thenStmt)
                + elsePart + ",\"location\":" + loc(1,1,5,1) + "}";
        return mapper.readValue(json, IfStmt.class);
    }

    public static Statement whileStmt(String condJson, Statement... body) throws Exception {
        StringBuilder stmtsJson = buildStmtsJson(body);
        String json = "{\"type\":\"WhileStmt\",\"condition\":" + condJson
                + ",\"body\":{\"type\":\"BlockStmt\",\"statements\":" + stmtsJson
                + ",\"location\":" + loc(2,1,5,1) + "},\"location\":" + loc(1,1,6,1) + "}";
        return mapper.readValue(json, WhileStmt.class);
    }

    public static Statement forStmt(Statement init, String condJson, Statement post, Statement... body) throws Exception {
        StringBuilder stmtsJson = buildStmtsJson(body);
        String initPart = (init != null) ? ",\"init\":" + stmtJson(init) : "";
        String postPart = (post != null) ? ",\"post\":" + stmtJson(post) : "";
        String json = "{\"type\":\"ForStmt\""
                + initPart + ",\"condition\":" + condJson
                + postPart
                + ",\"body\":{\"type\":\"BlockStmt\",\"statements\":" + stmtsJson
                + ",\"location\":" + loc(2,1,5,1) + "},\"location\":" + loc(1,1,6,1) + "}";
        return mapper.readValue(json, ForStmt.class);
    }

    public static Statement doWhileStmt(String condJson, Statement... body) throws Exception {
        StringBuilder stmtsJson = buildStmtsJson(body);
        String json = "{\"type\":\"DoWhileStmt\",\"condition\":" + condJson
                + ",\"body\":{\"type\":\"BlockStmt\",\"statements\":" + stmtsJson
                + ",\"location\":" + loc(2,1,5,1) + "},\"location\":" + loc(1,1,6,1) + "}";
        return mapper.readValue(json, DoWhileStmt.class);
    }

    public static Statement breakStmt() throws Exception {
        return mapper.readValue(
                "{\"type\":\"BreakStmt\",\"location\":" + loc(1,1,1,10) + "}",
                flowchart.ast.BreakStmt.class);
    }

    public static Statement continueStmt() throws Exception {
        return mapper.readValue(
                "{\"type\":\"ContinueStmt\",\"location\":" + loc(1,1,1,10) + "}",
                flowchart.ast.ContinueStmt.class);
    }

    // ─── Выражения (возвращают JSON-строки) ────────────────────────

    public static String intLit(int value) {
        return "{\"type\":\"IntLiteral\",\"value\":" + value + ",\"location\":" + loc(1,1,1,5) + "}";
    }

    public static String varExpr(String name) {
        return "{\"type\":\"VariableExpr\",\"name\":\"" + name + "\",\"location\":" + loc(1,1,1,5) + "}";
    }

    public static String binExpr(String left, String op, String right) {
        return "{\"type\":\"BinaryExpr\",\"operator\":\"" + op + "\","
                + "\"left\":" + left + ",\"right\":" + right
                + ",\"location\":" + loc(1,1,1,10) + "}";
    }

    public static String assignExpr(String left, String op, String right) {
        return "{\"type\":\"AssignmentExpr\",\"operator\":\"" + op + "\","
                + "\"left\":" + left + ",\"right\":" + right
                + ",\"location\":" + loc(1,1,1,10) + "}";
    }

    public static String unaryExpr(String operand, String op, boolean postfix) {
        return "{\"type\":\"UnaryExpr\",\"operator\":\"" + op + "\","
                + "\"operand\":" + operand + ",\"postfix\":" + postfix
                + ",\"location\":" + loc(1,1,1,5) + "}";
    }

    // ─── Вспомогательные ──────────────────────────────────────────

    private static String loc(int line, int col, int endLine, int endCol) {
        return "{\"line\":" + line + ",\"column\":" + col
                + ",\"endLine\":" + endLine + ",\"endColumn\":" + endCol + "}";
    }

    private static String stmtJson(Statement stmt) throws Exception {
        return mapper.writeValueAsString(stmt);
    }

    private static String funcJson(FunctionDecl func) throws Exception {
        return mapper.writeValueAsString(func);
    }

    private static StringBuilder buildStmtsJson(Statement[] stmts) throws Exception {
        StringBuilder sb = new StringBuilder("[");
        for (int i = 0; i < stmts.length; i++) {
            if (i > 0) sb.append(",");
            sb.append(stmtJson(stmts[i]));
        }
        sb.append("]");
        return sb;
    }
}