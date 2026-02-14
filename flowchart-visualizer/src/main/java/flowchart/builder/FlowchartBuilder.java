package flowchart.builder;

import flowchart.ast.*;
import flowchart.model.*;
import java.util.*;

public class FlowchartBuilder {

    private Map<String, FunctionDecl> functions = new HashMap<>();

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

        FlowchartNode body = buildStatement(func.getBody());

        if (body != null) {
            start.addNext(body);
        }

        return start;
    }

    private FlowchartNode buildStatement(Statement stmt) {
        if (stmt instanceof BlockStmt b) return buildBlock(b);
        if (stmt instanceof VariableDecl v) return buildVar(v);
        if (stmt instanceof ExprStmt e) return buildExpr(e);
        if (stmt instanceof IfStmt i) return buildIf(i);
        if (stmt instanceof WhileStmt w) return buildWhile(w);
        if (stmt instanceof ForStmt f) return buildFor(f);
        if (stmt instanceof ReturnStmt r) return buildReturn(r);
        if (stmt instanceof BreakStmt b) return new ConnectorNode("break");
        if (stmt instanceof ContinueStmt c) return new ConnectorNode("continue");
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
                    // Если prev это LoopStart, добавляем node как exitNode
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

    private FlowchartNode buildReturn(ReturnStmt r) {
        // return не создаёт узел - блок "конец" и так подразумевает завершение
        return null;
    }

    private FlowchartNode buildIf(IfStmt stmt) {
        DecisionNode decision = new DecisionNode(expr(stmt.getCondition()));
        decision.setAstLocation(toLocation(stmt.getLocation()));
        FlowchartNode thenNode = buildStatement(stmt.getThenBlock());
        decision.setTrueBranch(thenNode);
        FlowchartNode elseNode = null;
        if (stmt.getElseIfList() != null && !stmt.getElseIfList().isEmpty()) {
            elseNode = buildElseIfChain(stmt.getElseIfList(), stmt.getElseBlock());
        } else if (stmt.getElseBlock() != null) {
            elseNode = buildStatement(stmt.getElseBlock());
        }
        decision.setFalseBranch(elseNode);
        return decision;
    }

    private FlowchartNode buildElseIfChain(List<ElseIfClause> list, Statement finalElse) {
        ElseIfClause first = list.get(0);
        DecisionNode node = new DecisionNode(expr(first.getCondition()));
        node.setTrueBranch(buildStatement(first.getBlock()));
        List<ElseIfClause> rest = list.subList(1, list.size());
        node.setFalseBranch(rest.isEmpty()
                ? (finalElse != null ? buildStatement(finalElse) : null)
                : buildElseIfChain(rest, finalElse));
        return node;
    }

    private FlowchartNode buildWhile(WhileStmt stmt) {
        LoopStartNode start = new LoopStartNode(expr(stmt.getCondition()));
        start.setAstLocation(toLocation(stmt.getLocation()));

        // Тело цикла
        FlowchartNode body = buildStatement(stmt.getBody());
        start.setLoopBody(body);

        // Конец цикла - для стрелки назад
        LoopEndNode end = new LoopEndNode();
        end.setLoopStart(start);

        // Связываем конец тела с LoopEnd
        FlowchartNode lastBody = findExit(body);
        if (lastBody != null && lastBody != start) {
            lastBody.addNext(end);
        }

        // Выход из цикла будет добавлен в buildBlock через setExitNode

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

        // Терминатор - это выход
        if (node instanceof TerminalNode t && !t.isStart()) return node;

        // LoopStart - сам управляет выходом, возвращаем его
        if (node instanceof LoopStartNode) return node;

        // LoopEnd - это выход из тела цикла
        if (node instanceof LoopEndNode) return node;

        // Decision - имеет несколько ветвей
        if (node instanceof DecisionNode) return node;

        List<FlowchartNode> next = node.getNext();
        if (next.isEmpty()) return node;
        if (next.size() > 1) return node;

        return findExit(next.get(0), visited);
    }

    private String expr(Expression e) {
        if (e instanceof IntLiteral i) return String.valueOf(i.getValue());
        if (e instanceof VariableExpr v) return v.getName();
        if (e instanceof BinaryExpr b) return expr(b.getLeft()) + " " + b.getOp() + " " + expr(b.getRight());
        if (e instanceof UnaryExpr u) return u.getOp() + expr(u.getOperand());
        if (e instanceof AssignmentExpr a) return expr(a.getLeft()) + " " + a.getOp() + " " + expr(a.getRight());
        if (e instanceof CallExpr c) {
            StringBuilder sb = new StringBuilder(expr(c.getFunction()) + "(");
            for (int i = 0; i < c.getArguments().size(); i++) {
                if (i > 0) sb.append(", ");
                sb.append(expr(c.getArguments().get(i)));
            }
            return sb + ")";
        }
        return e.getClass().getSimpleName();
    }

    private flowchart.model.Location toLocation(ASTLocation a) {
        if (a == null) {
            // возвращаем фиктивную локацию, чтобы не падало
            return new flowchart.model.Location(0, 0, 0, 0);
        }

        return new flowchart.model.Location(
                a.getLine(),
                a.getColumn(),
                a.getEndLine(),
                a.getEndColumn()
        );
    }
}