package com.metrics.calculator;

import com.metrics.ast.*;
import com.metrics.model.FunctionMetrics;
import com.metrics.model.ProgramMetrics;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.List;

@Component
public class MetricsCalculator {

    public ProgramMetrics calculate(Program program) {
        ProgramMetrics result = new ProgramMetrics();
        List<FunctionMetrics> functions = new ArrayList<>();
        int globalVarCount = 0;

        for (Statement decl : program.getDeclarations()) {
            if (decl instanceof FunctionDecl fn) {
                functions.add(calculateFunction(fn));
            } else if (decl instanceof VariableDecl) {
                globalVarCount++;
            }
        }

        result.setFunctionCount(functions.size());
        result.setGlobalVarCount(globalVarCount);
        result.setFunctions(functions);
        return result;
    }

    private FunctionMetrics calculateFunction(FunctionDecl fn) {
        FunctionMetrics m = new FunctionMetrics();
        m.setFunctionName(fn.getName());
        m.setParameterCount(fn.getParameters() == null ? 0 : fn.getParameters().size());

        if (fn.getBody() != null) {
            // LOC = endLine - startLine + 1
            m.setLoc(calcLoc(fn));
            m.setCyclomaticComplexity(1 + countDecisionPoints(fn.getBody()));
            m.setMaxNestingDepth(calcMaxNesting(fn.getBody(), 0));
            m.setCallCount(countCalls(fn.getBody()));
            m.setReturnCount(countReturns(fn.getBody()));
            m.setGotoCount(countGotos(fn.getBody()));
        } else {
            m.setLoc(1);
            m.setCyclomaticComplexity(1);
        }
        return m;
    }

    // ── LOC ───────────────────────────────────────────────────────────────────
    private int calcLoc(FunctionDecl fn) {
        if (fn.getLocation() == null) return 1;
        int start = fn.getLocation().getLine();
        int end   = fn.getLocation().getEndLine();
        return Math.max(1, end - start + 1);
    }

    // ── Cyclomatic complexity (count decision points) ─────────────────────────
    private int countDecisionPoints(Statement stmt) {
        if (stmt == null) return 0;
        int count = 0;

        if (stmt instanceof IfStmt s) {
            count += 1; // if itself
            count += countDecisionPointsExpr(s.getCondition());
            count += countDecisionPoints(s.getThenBlock());
            if (s.getElseBlock() != null) count += countDecisionPoints(s.getElseBlock());

        } else if (stmt instanceof WhileStmt s) {
            count += 1;
            count += countDecisionPointsExpr(s.getCondition());
            count += countDecisionPoints(s.getBody());

        } else if (stmt instanceof ForStmt s) {
            count += 1;
            if (s.getCondition() != null) count += countDecisionPointsExpr(s.getCondition());
            count += countDecisionPoints(s.getBody());

        } else if (stmt instanceof DoWhileStmt s) {
            count += 1;
            count += countDecisionPointsExpr(s.getCondition());
            count += countDecisionPoints(s.getBody());

        } else if (stmt instanceof BlockStmt s) {
            if (s.getStatements() != null)
                for (Statement child : s.getStatements())
                    count += countDecisionPoints(child);

        } else if (stmt instanceof ReturnStmt s) {
            count += countDecisionPointsExpr(s.getValue());

        } else if (stmt instanceof ExprStmt s) {
            count += countDecisionPointsExpr(s.getExpression());

        } else if (stmt instanceof LabelStmt s) {
            count += countDecisionPoints(s.getStatement());
        }

        return count;
    }

    private int countDecisionPointsExpr(Expression expr) {
        if (expr == null) return 0;
        int count = 0;
        if (expr instanceof BinaryExpr e) {
            // && and || add a decision point
            if ("&&".equals(e.getOperator()) || "||".equals(e.getOperator())) count++;
            count += countDecisionPointsExpr(e.getLeft());
            count += countDecisionPointsExpr(e.getRight());
        } else if (expr instanceof UnaryExpr e) {
            count += countDecisionPointsExpr(e.getOperand());
        } else if (expr instanceof AssignmentExpr e) {
            count += countDecisionPointsExpr(e.getRight());
        } else if (expr instanceof CallExpr e) {
            if (e.getArguments() != null)
                for (Expression arg : e.getArguments())
                    count += countDecisionPointsExpr(arg);
        } else if (expr instanceof ArrayAccessExpr e) {
            count += countDecisionPointsExpr(e.getIndex());
        }
        return count;
    }

    // ── Max nesting depth ─────────────────────────────────────────────────────
    private int calcMaxNesting(Statement stmt, int depth) {
        if (stmt == null) return depth;

        if (stmt instanceof BlockStmt s) {
            int max = depth;
            if (s.getStatements() != null)
                for (Statement child : s.getStatements())
                    max = Math.max(max, calcMaxNesting(child, depth));
            return max;

        } else if (stmt instanceof IfStmt s) {
            int max = depth + 1;
            max = Math.max(max, calcMaxNesting(s.getThenBlock(), depth + 1));
            if (s.getElseBlock() != null)
                max = Math.max(max, calcMaxNesting(s.getElseBlock(), depth + 1));
            return max;

        } else if (stmt instanceof WhileStmt s) {
            return Math.max(depth + 1, calcMaxNesting(s.getBody(), depth + 1));

        } else if (stmt instanceof ForStmt s) {
            return Math.max(depth + 1, calcMaxNesting(s.getBody(), depth + 1));

        } else if (stmt instanceof DoWhileStmt s) {
            return Math.max(depth + 1, calcMaxNesting(s.getBody(), depth + 1));

        } else if (stmt instanceof LabelStmt s) {
            return calcMaxNesting(s.getStatement(), depth);
        }

        return depth;
    }

    // ── Call count ────────────────────────────────────────────────────────────
    private int countCalls(Statement stmt) {
        if (stmt == null) return 0;
        int count = 0;

        if (stmt instanceof BlockStmt s) {
            if (s.getStatements() != null)
                for (Statement child : s.getStatements()) count += countCalls(child);
        } else if (stmt instanceof IfStmt s) {
            count += countCallsExpr(s.getCondition());
            count += countCalls(s.getThenBlock());
            count += countCalls(s.getElseBlock());
        } else if (stmt instanceof WhileStmt s) {
            count += countCallsExpr(s.getCondition());
            count += countCalls(s.getBody());
        } else if (stmt instanceof ForStmt s) {
            count += countCallsExpr(s.getCondition());
            count += countCalls(s.getBody());
        } else if (stmt instanceof DoWhileStmt s) {
            count += countCalls(s.getBody());
            count += countCallsExpr(s.getCondition());
        } else if (stmt instanceof ExprStmt s) {
            count += countCallsExpr(s.getExpression());
        } else if (stmt instanceof ReturnStmt s) {
            count += countCallsExpr(s.getValue());
        } else if (stmt instanceof VariableDecl s) {
            count += countCallsExpr(s.getInitExpr());
        } else if (stmt instanceof LabelStmt s) {
            count += countCalls(s.getStatement());
        }
        return count;
    }

    private int countCallsExpr(Expression expr) {
        if (expr == null) return 0;
        int count = 0;
        if (expr instanceof CallExpr e) {
            count++;
            if (e.getArguments() != null)
                for (Expression arg : e.getArguments()) count += countCallsExpr(arg);
        } else if (expr instanceof BinaryExpr e) {
            count += countCallsExpr(e.getLeft()) + countCallsExpr(e.getRight());
        } else if (expr instanceof UnaryExpr e) {
            count += countCallsExpr(e.getOperand());
        } else if (expr instanceof AssignmentExpr e) {
            count += countCallsExpr(e.getLeft()) + countCallsExpr(e.getRight());
        } else if (expr instanceof ArrayAccessExpr e) {
            count += countCallsExpr(e.getArray()) + countCallsExpr(e.getIndex());
        }
        return count;
    }

    // ── Return count ──────────────────────────────────────────────────────────
    private int countReturns(Statement stmt) {
        if (stmt == null) return 0;
        if (stmt instanceof ReturnStmt) return 1;
        if (stmt instanceof BlockStmt s) {
            int count = 0;
            if (s.getStatements() != null)
                for (Statement child : s.getStatements()) count += countReturns(child);
            return count;
        }
        if (stmt instanceof IfStmt s)
            return countReturns(s.getThenBlock()) + countReturns(s.getElseBlock());
        if (stmt instanceof WhileStmt s) return countReturns(s.getBody());
        if (stmt instanceof ForStmt s)  return countReturns(s.getBody());
        if (stmt instanceof DoWhileStmt s) return countReturns(s.getBody());
        if (stmt instanceof LabelStmt s)   return countReturns(s.getStatement());
        return 0;
    }

    // ── Goto count ────────────────────────────────────────────────────────────
    private int countGotos(Statement stmt) {
        if (stmt == null) return 0;
        if (stmt instanceof GotoStmt) return 1;
        if (stmt instanceof BlockStmt s) {
            int count = 0;
            if (s.getStatements() != null)
                for (Statement child : s.getStatements()) count += countGotos(child);
            return count;
        }
        if (stmt instanceof IfStmt s)
            return countGotos(s.getThenBlock()) + countGotos(s.getElseBlock());
        if (stmt instanceof WhileStmt s)   return countGotos(s.getBody());
        if (stmt instanceof ForStmt s)     return countGotos(s.getBody());
        if (stmt instanceof DoWhileStmt s) return countGotos(s.getBody());
        if (stmt instanceof LabelStmt s)   return countGotos(s.getStatement());
        return 0;
    }
}
