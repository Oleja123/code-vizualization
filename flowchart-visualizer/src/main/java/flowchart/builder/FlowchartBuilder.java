package flowchart.builder;

import flowchart.ast.*;
import flowchart.model.*;
import java.util.*;

public class FlowchartBuilder {

    private Map<String, FunctionDecl> functions = new HashMap<>();

    // общий терминатор текущей функции
    private TerminalNode currentFunctionEnd;

    public FlowchartNode buildFromProgram(Program program) {
        for (Statement stmt : program.getDeclarations()) {
            if (stmt instanceof FunctionDecl func) {
                functions.put(func.getName(), func);
            }
        }

        FunctionDecl main = functions.get("main");
        if (main == null) throw new RuntimeException("main() not found");

        return buildFunction(main);
    }

    private FlowchartNode buildFunction(FunctionDecl func) {

        TerminalNode start = new TerminalNode(func.getName(), true);
        start.setAstLocation(toLocation(func.getLocation()));

        // создаём ОДИН общий конец
        currentFunctionEnd = new TerminalNode("конец", false);

        FlowchartNode body = buildStatement(func.getBody());

        if (body != null) {
            start.addNext(body);
            connectToEnd(body, new HashSet<>());
        } else {
            start.addNext(currentFunctionEnd);
        }

        return start;
    }

    /**
     * Гарантированно подключает все выходы графа к currentFunctionEnd
     */
    private void connectToEnd(FlowchartNode node, Set<FlowchartNode> visited) {
        if (node == null || visited.contains(node)) return;
        visited.add(node);

        // return всегда ведёт в конец
        if (node instanceof ProcessNode p &&
                p.getLabel() != null &&
                p.getLabel().startsWith("return")) {

            node.getNext().clear();
            node.addNext(currentFunctionEnd);
            return;
        }

        // если узел уже конечный терминатор
        if (node instanceof TerminalNode t && !t.isStart()) {
            return;
        }

        // Decision
        if (node instanceof DecisionNode decision) {

            connectToEnd(decision.getTrueBranch(), visited);
            connectToEnd(decision.getFalseBranch(), visited);

            if (decision.getNext().isEmpty()) {
                decision.addNext(currentFunctionEnd);
            } else {
                for (FlowchartNode n : decision.getNext()) {
                    connectToEnd(n, visited);
                }
            }
            return;
        }

        // LoopStart
        if (node instanceof LoopStartNode loop) {

            connectToEnd(loop.getLoopBody(), visited);
            connectToEnd(loop.getExitNode(), visited);

            if (loop.getExitNode() == null) {
                loop.setExitNode(currentFunctionEnd);
            }

            return;
        }

        // если нет продолжения — это лист
        if (node.getNext().isEmpty()) {
            node.addNext(currentFunctionEnd);
            return;
        }

        for (FlowchartNode n : node.getNext()) {
            connectToEnd(n, visited);
        }
    }

    private FlowchartNode buildStatement(Statement stmt) {
        if (stmt instanceof BlockStmt b) return buildBlock(b);
        if (stmt instanceof VariableDecl v) return buildVar(v);
        if (stmt instanceof ExprStmt e) return buildExpr(e);
        if (stmt instanceof IfStmt i) return buildIf(i);
        if (stmt instanceof WhileStmt w) return buildWhile(w);
        if (stmt instanceof ForStmt f) return buildFor(f);
        if (stmt instanceof ReturnStmt r) return buildReturn(r);
        if (stmt instanceof BreakStmt) return new ConnectorNode("break");
        if (stmt instanceof ContinueStmt) return new ConnectorNode("continue");
        throw new RuntimeException("Unknown stmt: " + stmt);
    }

    private FlowchartNode buildBlock(BlockStmt block) {

        FlowchartNode first = null;
        FlowchartNode prev = null;

        for (Statement stmt : block.getStatements()) {

            FlowchartNode node = buildStatement(stmt);
            if (node == null) continue;

            if (first == null) first = node;

            if (prev != null) {

                FlowchartNode last = findExit(prev);

                if (last != null) {

                    if (prev instanceof LoopStartNode loop) {
                        loop.setExitNode(node);
                    } else {
                        last.addNext(node);
                    }
                }
            }

            prev = node;
        }

        return first;
    }

    private FlowchartNode buildVar(VariableDecl d) {
        String label = d.getVarType() + " " + d.getName();
        if (d.getInitExpr() != null) label += " = " + expr(d.getInitExpr());

        ProcessNode p = new ProcessNode(label);
        p.setAstLocation(toLocation(d.getLocation()));
        return p;
    }

    private FlowchartNode buildExpr(ExprStmt e) {
        ProcessNode p = new ProcessNode(expr(e.getExpression()));
        p.setAstLocation(toLocation(e.getLocation()));
        return p;
    }

    // return теперь полноценный узел
    private FlowchartNode buildReturn(ReturnStmt r) {
        String label = r.getValue() != null
                ? "return " + expr(r.getValue())
                : "return";

        ProcessNode node = new ProcessNode(label);
        node.setAstLocation(toLocation(r.getLocation()));
        return node;
    }

    private FlowchartNode buildIf(IfStmt stmt) {

        DecisionNode decision = new DecisionNode(expr(stmt.getCondition()));
        decision.setAstLocation(toLocation(stmt.getLocation()));

        FlowchartNode thenNode = buildStatement(stmt.getThenBlock());
        FlowchartNode elseNode = stmt.getElseBlock() != null
                ? buildStatement(stmt.getElseBlock())
                : null;

        decision.setTrueBranch(thenNode);
        decision.setFalseBranch(elseNode);

        return decision;
    }

    private FlowchartNode buildWhile(WhileStmt stmt) {

        LoopStartNode start = new LoopStartNode(expr(stmt.getCondition()));
        start.setAstLocation(toLocation(stmt.getLocation()));

        FlowchartNode body = buildStatement(stmt.getBody());
        start.setLoopBody(body);

        LoopEndNode end = new LoopEndNode();
        end.setLoopStart(start);

        FlowchartNode lastBody = findExit(body);
        if (lastBody != null && lastBody != start) {
            lastBody.addNext(end);
        }

        return start;
    }

    private FlowchartNode buildFor(ForStmt stmt) {

        FlowchartNode init = stmt.getInit() != null ? buildStatement(stmt.getInit()) : null;

        LoopStartNode start = new LoopStartNode(
                stmt.getCondition() != null ? expr(stmt.getCondition()) : "true");

        FlowchartNode body = buildStatement(stmt.getBody());
        start.setLoopBody(body);

        FlowchartNode post = stmt.getPost() != null ? buildStatement(stmt.getPost()) : null;

        LoopEndNode end = new LoopEndNode();
        end.setLoopStart(start);

        FlowchartNode lastBody = findExit(body);

        if (lastBody != null && lastBody != start) {
            if (post != null) {
                lastBody.addNext(post);
                post.addNext(end);
            } else {
                lastBody.addNext(end);
            }
        }

        if (init != null) {
            init.addNext(start);
            return init;
        }

        return start;
    }

    private FlowchartNode findExit(FlowchartNode node) {
        return findExit(node, new HashSet<>());
    }

    private FlowchartNode findExit(FlowchartNode node, Set<FlowchartNode> visited) {

        if (node == null || visited.contains(node)) return null;
        visited.add(node);

        if (node instanceof TerminalNode t && !t.isStart()) return node;
        if (node instanceof LoopStartNode) return node;
        if (node instanceof LoopEndNode) return node;
        if (node instanceof DecisionNode) return node;

        List<FlowchartNode> next = node.getNext();

        if (next.isEmpty()) return node;
        if (next.size() > 1) return node;

        return findExit(next.get(0), visited);
    }

    private String expr(Expression e) {

        if (e == null) return "?";

        if (e instanceof IntLiteral i) return String.valueOf(i.getValue());
        if (e instanceof VariableExpr v) return v.getName();
        if (e instanceof BinaryExpr b)
            return expr(b.getLeft()) + " " + b.getOp() + " " + expr(b.getRight());
        if (e instanceof UnaryExpr u)
            return u.isPostfix()
                    ? expr(u.getOperand()) + u.getOp()
                    : u.getOp() + expr(u.getOperand());
        if (e instanceof AssignmentExpr a)
            return expr(a.getLeft()) + " " + a.getOp() + " " + expr(a.getRight());

        if (e instanceof CallExpr c) {
            StringBuilder sb = new StringBuilder(c.getFunctionName() + "(");
            if (c.getArguments() != null) {
                for (int i = 0; i < c.getArguments().size(); i++) {
                    if (i > 0) sb.append(", ");
                    sb.append(expr(c.getArguments().get(i)));
                }
            }
            return sb.append(")").toString();
        }

        return e.getClass().getSimpleName();
    }

    private flowchart.model.Location toLocation(ASTLocation a) {
        return new flowchart.model.Location(
                a.getLine(),
                a.getColumn(),
                a.getEndLine(),
                a.getEndColumn()
        );
    }
}
