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

        TerminalNode end = new TerminalNode("конец", false);

        FlowchartNode body = buildStatement(func.getBody());

        if (body != null) {
            start.addNext(body);
            attachEnd(body, end, new HashSet<>());
        } else {
            start.addNext(end);
        }

        return start;
    }

    /**
     * Рекурсивно обходит граф и цепляет `end` ко всем листьям.
     */
    private void attachEnd(FlowchartNode node, TerminalNode end, Set<FlowchartNode> visited) {
        if (node == null || visited.contains(node)) return;
        if (node == end) return;
        visited.add(node);

        // return-узел → привязываем конец
        if (node instanceof ProcessNode p
                && p.getLabel() != null
                && p.getLabel().startsWith("return")) {
            if (p.getNext().isEmpty()) {
                p.addNext(end);
            }
            return;
        }

        // break/continue коннекторы — не трогаем, они управляются циклом
        if (node instanceof ConnectorNode) {
            return;
        }

        if (node instanceof LoopStartNode loop) {
            // Обходим тело цикла
            attachEnd(loop.getLoopBody(), end, visited);
            // EXIT цикла — НЕ трогаем: exitNode должен быть следующим узлом после цикла,
            // который будет подключён через linkNodes в buildBlock.
            // Если exitNode уже установлен — обходим его продолжение.
            if (loop.getExitNode() != null) {
                attachEnd(loop.getExitNode(), end, visited);
            }
            // Если exitNode == null — его подключит linkNodes или он останется null
            // (attachEnd не должен сам ставить end как exitNode!)
            return;
        }

        if (node instanceof DoWhileNode doWhile) {
            attachEnd(doWhile.getLoopBody(), end, visited);
            if (doWhile.getExitNode() != null) {
                attachEnd(doWhile.getExitNode(), end, visited);
            } else {
                // do-while последний оператор — конец идёт сразу после него
                doWhile.setExitNode(end);
            }
            return;
        }

        if (node instanceof DecisionNode decision) {
            attachEnd(decision.getTrueBranch(), end, visited);
            attachEnd(decision.getFalseBranch(), end, visited);
            // Sanitize NEXT: remove any entries that are branches or null (guard against builder bugs)
            decision.getNext().removeIf(n ->
                    n == null ||
                            n == decision.getTrueBranch() ||
                            n == decision.getFalseBranch());
            for (FlowchartNode n : new ArrayList<>(decision.getNext())) {
                attachEnd(n, end, visited);
            }
            return;
        }

        if (node.getNext().isEmpty()) {
            node.addNext(end);
            return;
        }

        for (FlowchartNode next : new ArrayList<>(node.getNext())) {
            attachEnd(next, end, visited);
        }
    }

    // ──────────────────────────────────────────────────────────
    //  Statement builders
    // ──────────────────────────────────────────────────────────

    private FlowchartNode buildStatement(Statement stmt) {
        if (stmt instanceof BlockStmt b)    return buildBlock(b);
        if (stmt instanceof VariableDecl v) return buildVar(v);
        if (stmt instanceof ExprStmt e)     return buildExpr(e);
        if (stmt instanceof IfStmt i)       return buildIf(i);
        if (stmt instanceof WhileStmt w)    return buildWhile(w);
        if (stmt instanceof ForStmt f)      return buildFor(f);
        if (stmt instanceof DoWhileStmt d)  return buildDoWhile(d);
        if (stmt instanceof ReturnStmt r)   return buildReturn(r);
        if (stmt instanceof BreakStmt)      return new ConnectorNode("break");
        if (stmt instanceof ContinueStmt)   return new ConnectorNode("continue");
        throw new RuntimeException("Unknown stmt: " + stmt);
    }

    private FlowchartNode buildBlock(BlockStmt block) {
        FlowchartNode first = null;
        FlowchartNode prev  = null;

        for (Statement stmt : block.getStatements()) {
            FlowchartNode node = buildStatement(stmt);
            if (node == null) continue;

            if (first == null) first = node;

            if (prev != null) {
                linkNodes(prev, node);
            }
            prev = node;
        }

        return first;
    }

    /**
     * Соединяет конец предыдущего узла с началом следующего.
     */
    private void linkNodes(FlowchartNode prev, FlowchartNode next) {
        if (prev instanceof LoopStartNode loop) {
            if (loop.getExitNode() == null) {
                loop.setExitNode(next);
            } else {
                linkNodes(loop.getExitNode(), next);
            }
            return;
        }

        if (prev instanceof DoWhileNode doWhile) {
            if (doWhile.getExitNode() == null) {
                doWhile.setExitNode(next);
            } else {
                linkNodes(doWhile.getExitNode(), next);
            }
            return;
        }

        if (prev instanceof DecisionNode decision) {
            // Не добавлять в NEXT то, что уже является веткой
            if (next == decision.getTrueBranch()) return;
            if (next == decision.getFalseBranch()) return;
            if (next != null && next != decision.getTrueBranch() && next != decision.getFalseBranch()
                    && !decision.getNext().contains(next)) {
                decision.addNext(next);
            }
            return;
        }

        Set<FlowchartNode> visited = new HashSet<>();
        linkLeaves(prev, next, visited);
    }

    private void linkLeaves(FlowchartNode node, FlowchartNode next, Set<FlowchartNode> visited) {
        if (node == null || visited.contains(node) || node == next) return;
        visited.add(node);

        if (node instanceof LoopStartNode loop) {
            if (loop.getExitNode() == null) {
                loop.setExitNode(next);
            } else {
                linkLeaves(loop.getExitNode(), next, visited);
            }
            return;
        }

        if (node instanceof DoWhileNode doWhile) {
            if (doWhile.getExitNode() == null) {
                doWhile.setExitNode(next);
            } else {
                linkLeaves(doWhile.getExitNode(), next, visited);
            }
            return;
        }

        if (node instanceof DecisionNode decision) {
            if (next == decision.getTrueBranch()) return;
            if (next == decision.getFalseBranch()) return;
            if (next != null && next != decision.getTrueBranch() && next != decision.getFalseBranch()
                    && !decision.getNext().contains(next)) {
                decision.addNext(next);
            }
            return;
        }

        // break/continue — не подключаем к следующему узлу блока
        if (node instanceof ConnectorNode) {
            return;
        }

        if (node.getNext().isEmpty()) {
            node.addNext(next);
            return;
        }

        for (FlowchartNode child : new ArrayList<>(node.getNext())) {
            if (child instanceof LoopEndNode) continue;
            linkLeaves(child, next, visited);
        }
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

        // Хвосты тела → LoopEndNode
        if (body != null) {
            Set<FlowchartNode> visited = new HashSet<>();
            linkLeaves(body, end, visited);
        }

        // LoopEndNode → обратно к началу (только один раз!)
        if (!end.getNext().contains(start)) {
            end.addNext(start);
        }

        // exitNode будет установлен через linkNodes в buildBlock,
        // когда после while идёт следующий оператор.
        // Здесь намеренно НЕ устанавливаем exitNode.

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

        if (body != null) {
            Set<FlowchartNode> visited = new HashSet<>();
            if (post != null) {
                linkLeaves(body, post, visited);
                post.addNext(end);
            } else {
                linkLeaves(body, end, visited);
            }
        }

        // Только один переход назад
        end.addNext(start);

        if (init != null) {
            init.addNext(start);
            return init;
        }

        return start;
    }

    // ──────────────────────────────────────────────────────────
    //  Helpers
    // ──────────────────────────────────────────────────────────

    private String expr(Expression e) {
        if (e == null) return "?";
        if (e instanceof IntLiteral i)       return String.valueOf(i.getValue());
        if (e instanceof VariableExpr v)     return v.getName();
        if (e instanceof ArrayAccessExpr a)  return expr(a.getArray()) + "[" + expr(a.getIndex()) + "]";
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
                a.getLine(), a.getColumn(),
                a.getEndLine(), a.getEndColumn());
    }

    private FlowchartNode buildDoWhile(DoWhileStmt stmt) {
        DoWhileNode node = new DoWhileNode(expr(stmt.getCondition()));
        node.setAstLocation(toLocation(stmt.getLocation()));

        FlowchartNode body = buildStatement(stmt.getBody());
        node.setLoopBody(body);

        LoopEndNode end = new LoopEndNode();
        end.setLoopStart(node); // reuse LoopEndNode pointing back to DoWhileNode

        // Connect body leaves → LoopEndNode → back to condition (DoWhileNode renders this as back-arrow)
        if (body != null) {
            Set<FlowchartNode> visited = new HashSet<>();
            linkLeaves(body, end, visited);
        }

        // LoopEndNode → back to DoWhileNode (the renderer uses this to know body is done)
        if (!end.getNext().contains(node)) {
            end.addNext(node);
        }

        // exitNode will be set by linkNodes when the statement after do-while is connected
        return node;
    }
}