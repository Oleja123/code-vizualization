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

        // Конец — всегда один, всегда самый последний
        TerminalNode end = new TerminalNode("конец", false);

        // Строим тело
        FlowchartNode body = buildStatement(func.getBody());

        // Ищем «хвост» линейной цепочки (последний узел перед концом)
        // и подцепляем конец к нему
        if (body != null) {
            start.addNext(body);
            attachEnd(body, end, new HashSet<>());
        } else {
            start.addNext(end);
        }

        return start;
    }

    /**
     * Рекурсивно обходит граф и цепляет `end` ко всем
     * «листьям» — узлам у которых нет исходящих рёбер.
     *
     * Правило одно: листом считается узел, у которого next пуст,
     * и при этом это не сам терминал конца.
     */
    private void attachEnd(FlowchartNode node, TerminalNode end, Set<FlowchartNode> visited) {
        if (node == null || visited.contains(node)) return;
        if (node == end) return;
        visited.add(node);

        // ProcessNode с return — подключаем к концу
        if (node instanceof ProcessNode p
                && p.getLabel() != null
                && p.getLabel().startsWith("return")) {
            if (p.getNext().isEmpty()) {
                p.addNext(end);
            }
            return;
        }

        // Листовой узел (нет next) — подключаем к концу
        if (node.getNext().isEmpty()) {
            node.addNext(end);
            return;
        }

        // Рекурсия по всем дочерним
        for (FlowchartNode next : new ArrayList<>(node.getNext())) {
            attachEnd(next, end, visited);
        }

        // DecisionNode: дополнительно обходим ветки
        if (node instanceof DecisionNode decision) {
            attachEnd(decision.getTrueBranch(), end, visited);
            attachEnd(decision.getFalseBranch(), end, visited);
        }

        // LoopStartNode: обходим тело и exitNode
        if (node instanceof LoopStartNode loop) {
            attachEnd(loop.getLoopBody(), end, visited);
            // Для exitNode тоже нужно вызвать attachEnd
            if (loop.getExitNode() != null && loop.getExitNode() != end) {
                attachEnd(loop.getExitNode(), end, visited);
            } else if (loop.getExitNode() == null) {
                loop.setExitNode(end);
            }
        }
    }

    // ──────────────────────────────────────────────────────────
    //  Statement builders
    // ──────────────────────────────────────────────────────────

    private FlowchartNode buildStatement(Statement stmt) {
        if (stmt instanceof BlockStmt b)   return buildBlock(b);
        if (stmt instanceof VariableDecl v) return buildVar(v);
        if (stmt instanceof ExprStmt e)    return buildExpr(e);
        if (stmt instanceof IfStmt i)      return buildIf(i);
        if (stmt instanceof WhileStmt w)   return buildWhile(w);
        if (stmt instanceof ForStmt f)     return buildFor(f);
        if (stmt instanceof ReturnStmt r)  return buildReturn(r);
        if (stmt instanceof BreakStmt)     return new ConnectorNode("break");
        if (stmt instanceof ContinueStmt)  return new ConnectorNode("continue");
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
                FlowchartNode last = findTail(prev);
                if (last != null) {
                    // Для LoopStartNode нужно подключить следующий узел как exitNode
                    if (last instanceof LoopStartNode loop) {
                        // Если у цикла еще нет exitNode, устанавливаем его
                        if (loop.getExitNode() == null) {
                            loop.setExitNode(node);
                        } else {
                            // Если exitNode уже установлен, добавляем следующий узел к нему
                            FlowchartNode exitTail = findTail(loop.getExitNode());
                            if (exitTail != null && exitTail.getNext().isEmpty()) {
                                exitTail.addNext(node);
                            }
                        }
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

        FlowchartNode lastBody = findTail(body);
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

        FlowchartNode lastBody = findTail(body);
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

    // ──────────────────────────────────────────────────────────
    //  Helpers
    // ──────────────────────────────────────────────────────────

    /**
     * Находит «хвостовой» узел линейной цепочки.
     * Останавливается на LoopStart/LoopEnd/Decision — они сами управляют ветвлением.
     */
    private FlowchartNode findTail(FlowchartNode node) {
        return findTail(node, new HashSet<>());
    }

    private FlowchartNode findTail(FlowchartNode node, Set<FlowchartNode> visited) {
        if (node == null || visited.contains(node)) return null;
        visited.add(node);

        if (node instanceof TerminalNode t && !t.isStart()) return node;
        if (node instanceof LoopStartNode) return node;
        if (node instanceof LoopEndNode)   return node;
        if (node instanceof DecisionNode)  return node;

        List<FlowchartNode> next = node.getNext();
        if (next.isEmpty())  return node;
        if (next.size() > 1) return node;

        return findTail(next.get(0), visited);
    }

    private String expr(Expression e) {
        if (e == null) return "?";
        if (e instanceof IntLiteral i)    return String.valueOf(i.getValue());
        if (e instanceof VariableExpr v)  return v.getName();
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
}