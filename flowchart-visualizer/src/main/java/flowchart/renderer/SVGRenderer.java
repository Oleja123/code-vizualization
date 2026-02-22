package flowchart.renderer;

import flowchart.model.*;
import java.util.*;

public class SVGRenderer {

    private static final double PROCESS_WIDTH    = 220;
    private static final double PROCESS_HEIGHT   = 70;
    private static final double TERMINAL_WIDTH   = 220;
    private static final double TERMINAL_HEIGHT  = 60;
    private static final double DECISION_WIDTH   = 220;
    private static final double DECISION_HEIGHT  = 120;

    private static final double VERTICAL_SPACING          = 80;
    private static final double HORIZONTAL_SPACING        = 260;
    private static final double BREAK_HORIZONTAL_SPACING  = 130;

    private static final double BACK_ARROW_MARGIN = 40;

    private StringBuilder svg;
    private Set<FlowchartNode> rendered;
    private Map<DecisionNode, double[]> breakGeometry;
    private Map<DecisionNode, double[]> continueGeometry;
    private double maxY = 0;
    private double minX = Double.MAX_VALUE;
    private double maxX = Double.MIN_VALUE;

    private TerminalNode endNode;
    private double endArrowFromX;
    private double endArrowFromY;

    private double currentLoopReturnRightX = Double.MAX_VALUE;
    private double lastBodyBlockX = Double.NaN;

    private List<Object[]> deferredDoWhileContinueCols;

    public String render(FlowchartNode start) {
        svg      = new StringBuilder();
        rendered = new HashSet<>();
        breakGeometry = new HashMap<>();
        continueGeometry = new HashMap<>();
        maxY = 0;
        minX = Double.MAX_VALUE;
        maxX = Double.MIN_VALUE;
        endNode = null;
        endArrowFromX = 700;
        endArrowFromY = 0;
        currentLoopReturnRightX = Double.MAX_VALUE;
        lastBodyBlockX = Double.NaN;
        deferredDoWhileContinueCols = new ArrayList<>();

        renderNode(start, 700, 100, null);

        double padding = 60;

        if (endNode != null) {
            double endX = endArrowFromX;
            double endY = endArrowFromY + VERTICAL_SPACING;
            arrow(endX, endArrowFromY, endX, endY);
            renderTerminal(endNode, endX, endY);
            updateMaxY(endY + TERMINAL_HEIGHT);
        }

        double svgW = Math.max(maxX - minX + padding * 2, 400);
        double svgH = maxY + padding;

        StringBuilder out = new StringBuilder();
        out.append("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n");
        out.append(String.format(Locale.US,
                "<svg xmlns=\"http://www.w3.org/2000/svg\" " +
                        "width=\"100%%\" height=\"100%%\" " +
                        "viewBox=\"%.1f 0 %.1f %.1f\" " +
                        "preserveAspectRatio=\"xMidYMin meet\">\n",
                minX - padding, svgW, svgH));

        out.append("<defs>\n");
        out.append("<marker id=\"arrow\" markerWidth=\"10\" markerHeight=\"10\" refX=\"9\" refY=\"5\" orient=\"auto\">\n");
        out.append("<path d=\"M0,0 L10,5 L0,10 z\" fill=\"black\"/>\n");
        out.append("</marker>\n");
        out.append("<style>\n");
        out.append(".shape  { fill: white; stroke: black; stroke-width: 2; }\n");
        out.append(".line   { stroke: black; stroke-width: 2; fill: none; }\n");
        out.append(".arrow  { stroke: black; stroke-width: 2; fill: none; marker-end: url(#arrow); }\n");
        out.append(".text   { font-family: Arial; font-size: 13px; text-anchor: middle; dominant-baseline: middle; }\n");
        out.append(".label  { font-family: Arial; font-size: 11px; fill: #333; }\n");
        out.append("</style>\n");
        out.append("</defs>\n");

        out.append(svg);
        out.append("</svg>");
        return out.toString();
    }

    private void trackX(double x) {
        if (x < minX) minX = x;
        if (x > maxX) maxX = x;
    }

    private void updateMaxY(double y) {
        if (y > maxY) maxY = y;
    }

    private void renderNode(FlowchartNode node, double x, double y, FlowchartNode stopBefore) {
        if (node == null || rendered.contains(node)) return;
        if (node == stopBefore) return;

        if (node instanceof TerminalNode t && !t.isStart()) {
            endNode = t;
            return;
        }
        if (node instanceof LoopEndNode) return;

        rendered.add(node);
        node.setPosition(x, y);

        trackX(x - PROCESS_WIDTH / 2);
        trackX(x + PROCESS_WIDTH / 2);

        switch (node.getType()) {
            case TERMINAL -> {
                renderTerminal(node, x, y);
                updateMaxY(y + TERMINAL_HEIGHT);
                renderLinearNext(node, x, y, null);
            }
            case PROCESS -> {
                renderProcess(node, x, y);
                updateMaxY(y + PROCESS_HEIGHT);
                renderLinearNext(node, x, y, stopBefore);
            }
            case DECISION -> renderDecision((DecisionNode) node, x, y, stopBefore);
            case LOOP_START -> renderLoop((LoopStartNode) node, x, y, stopBefore);
            case LOOP_END  -> { }
            case DO_WHILE  -> renderDoWhile((DoWhileNode) node, x, y, stopBefore);
        }
    }

    private void renderTerminal(FlowchartNode node, double x, double y) {
        double w = TERMINAL_WIDTH, h = TERMINAL_HEIGHT;
        node.setSize(w, h);
        svg.append(String.format(Locale.US,
                "<ellipse class=\"shape\" cx=\"%.1f\" cy=\"%.1f\" rx=\"%.1f\" ry=\"%.1f\"/>\n",
                x, y + h / 2, w / 2, h / 2));
        text(node.getLabel(), x, y + h / 2);
        trackX(x - w / 2);
        trackX(x + w / 2);
    }

    private void renderProcess(FlowchartNode node, double x, double y) {
        node.setSize(PROCESS_WIDTH, PROCESS_HEIGHT);
        svg.append(String.format(Locale.US,
                "<rect class=\"shape\" x=\"%.1f\" y=\"%.1f\" width=\"%.1f\" height=\"%.1f\"/>\n",
                x - PROCESS_WIDTH / 2, y, PROCESS_WIDTH, PROCESS_HEIGHT));
        text(node.getLabel(), x, y + PROCESS_HEIGHT / 2);
    }

    private void renderDecision(DecisionNode node, double x, double y, FlowchartNode stopBefore) {
        double w = DECISION_WIDTH, h = DECISION_HEIGHT;
        double halfW = w / 2, halfH = h / 2;
        node.setSize(w, h);
        drawDiamond(x, y, w, h);
        text(node.getLabel(), x, y + halfH);

        double branchY = y + h + VERTICAL_SPACING;
        double leftX   = x - HORIZONTAL_SPACING;
        double rightX  = x + HORIZONTAL_SPACING;
        trackX(leftX - PROCESS_WIDTH / 2);
        trackX(rightX + PROCESS_WIDTH / 2);

        double leftBottom  = branchY;
        double rightBottom = branchY;

        if (node.getTrueBranch() != null) {
            line(x - halfW, y + halfH, leftX, y + halfH);
            line(leftX, y + halfH, leftX, branchY - 5);
            arrow(leftX, branchY - 5, leftX, branchY);
            labelText("ДА", x - halfW - 30, y + halfH - 10);
            renderNode(node.getTrueBranch(), leftX, branchY, stopBefore);
            leftBottom = branchY + branchHeight(node.getTrueBranch(), stopBefore);
        }

        if (node.getFalseBranch() != null) {
            line(x + halfW, y + halfH, rightX, y + halfH);
            line(rightX, y + halfH, rightX, branchY - 5);
            arrow(rightX, branchY - 5, rightX, branchY);
            labelText("НЕТ", x + halfW + 10, y + halfH - 10);
            renderNode(node.getFalseBranch(), rightX, branchY, stopBefore);
            rightBottom = branchY + branchHeight(node.getFalseBranch(), stopBefore);
        } else {
            rightBottom = leftBottom;
        }

        double mergeY = Math.max(leftBottom, rightBottom) + VERTICAL_SPACING;

        if (node.getTrueBranch() != null) {
            line(leftX, leftBottom, leftX, mergeY);
            line(leftX, mergeY, x, mergeY);
        }
        if (node.getFalseBranch() != null) {
            line(rightX, rightBottom, rightX, mergeY);
            line(rightX, mergeY, x, mergeY);
        } else {
            double noElseRightX = x + HORIZONTAL_SPACING;
            line(x + halfW, y + halfH, noElseRightX, y + halfH);
            labelText("НЕТ", x + halfW + 10, y + halfH - 10);
            line(noElseRightX, y + halfH, noElseRightX, mergeY);
            line(noElseRightX, mergeY, x, mergeY);
            trackX(noElseRightX + 5);
        }

        node.setSize(w, mergeY - y);
        updateMaxY(mergeY);

        List<FlowchartNode> nexts = node.getNext();
        if (!nexts.isEmpty()) {
            double nextY = mergeY + VERTICAL_SPACING;
            arrow(x, mergeY, x, nextY - 5);
            for (FlowchartNode next : nexts) {
                renderNode(next, x, nextY, stopBefore);
            }
        } else {
            TerminalNode terminal = findEndTerminal(node);
            if (terminal != null) {
                endNode = terminal;
            }
            endArrowFromX = x;
            endArrowFromY = mergeY;
        }
    }

    private double branchHeight(FlowchartNode node, FlowchartNode stopBefore) {
        if (node == null || node == stopBefore) return 0;
        if (node instanceof TerminalNode)        return 0;
        if (node instanceof LoopEndNode)         return 0;

        double h = node.getHeight();

        if (node instanceof ProcessNode) {
            for (FlowchartNode next : node.getNext()) {
                if (next == stopBefore)           break;
                if (next instanceof TerminalNode) break;
                if (next instanceof LoopEndNode)  break;
                h += VERTICAL_SPACING + branchHeight(next, stopBefore);
                break;
            }
        }
        if (node instanceof LoopStartNode loop) {
            h = computeLoopTotalHeight(loop, stopBefore);
        }
        return h;
    }

    private double computeLoopTotalHeight(LoopStartNode loop, FlowchartNode stopBefore) {
        return loop.getHeight();
    }

    // ── LOOP ──────────────────────────────────────────────────────────────────

    private void renderLoop(LoopStartNode node, double x, double y, FlowchartNode stopBefore) {
        double w = DECISION_WIDTH, h = DECISION_HEIGHT;
        double halfW = w / 2, halfH = h / 2;

        drawDiamond(x, y, w, h);
        text(node.getLabel(), x, y + halfH);

        double rightX  = x + HORIZONTAL_SPACING;
        double branchY = y + h + VERTICAL_SPACING;

        line(x + halfW, y + halfH, rightX, y + halfH);
        arrow(rightX, y + halfH, rightX, branchY - 5);
        labelText("ДА", x + halfW + 10, y + halfH - 10);

        currentLoopReturnRightX = Double.MAX_VALUE;

        // Render full body (including tail blocks via renderBreakTailChain)
        List<FlowchartNode> bodyChain = collectBodyChain(node.getLoopBody(), node);
        double bodyEndY = renderLoopBodyChain(bodyChain, rightX, branchY, node);
        if (bodyEndY < 0) bodyEndY = -bodyEndY;

        // Compute returnRightX using only this loop's body X-extent,
        // NOT global maxX (which may include sibling columns of an outer if).
        double bodyMaxX = rightX + PROCESS_WIDTH / 2;
        for (double[] g : breakGeometry.values()) {
            bodyMaxX = Math.max(bodyMaxX, g[1] + PROCESS_WIDTH / 2); // netColX right edge
        }
        double returnRightX = bodyMaxX + BACK_ARROW_MARGIN;
        currentLoopReturnRightX = returnRightX;

        double diamondBottom = y + h;
        labelText("НЕТ", x + 8, diamondBottom + 15);

        double maxCornerY = bodyEndY + VERTICAL_SPACING / 2;
        for (double[] g : breakGeometry.values()) {
            double netColEndY = g[3];
            double daColEndY  = g[2];
            boolean isContinueDec = (g.length > 4 && g[4] == 1.0);
            maxCornerY = Math.max(maxCornerY, netColEndY + VERTICAL_SPACING / 2);
            if (isContinueDec) maxCornerY = Math.max(maxCornerY, daColEndY + VERTICAL_SPACING / 2);
        }

        double exitY = maxCornerY + VERTICAL_SPACING / 2;

        for (double[] g : breakGeometry.values()) {
            double daColX     = g[0];
            double netColX    = g[1];
            double daColEndY  = g[2];
            double netColEndY = g[3];
            boolean isContinueDec = (g.length > 4 && g[4] == 1.0);
            double tipY      = (g.length > 5) ? g[5] : 0;
            double colStartY = (g.length > 6) ? g[6] : tipY + DECISION_HEIGHT / 2 + VERTICAL_SPACING;

            boolean daBare  = (daColEndY  <= colStartY + 1);
            boolean netBare = (netColEndY <= colStartY + 1);

            // Pre-draw bare DA arrow only for break decisions where DA bare but NET has blocks
            if (daBare && !netBare && !isContinueDec) {
                arrow(daColX, tipY, daColX, colStartY);
            }

            double daLineStartY  = daBare  ? tipY    : daColEndY;
            double netLineStartY = netBare ? colStartY : netColEndY;

            if (isContinueDec) {
                // ── CONTINUE DECISION routing ──────────────────────────────
                // daColX=LEFT(continue branch), netColX=RIGHT(normal/bare)
                double contTargetY = y - VERTICAL_SPACING / 2;

                if (!daBare) {
                    // DA has blocks: right side of last block → horizontal to netColX → up to tipY
                    double blockRightX = daColX + PROCESS_WIDTH / 2;
                    double blockBottomY = daColEndY; // bottom of last block
                    // horizontal from right edge of last block to netColX, at block bottom level
                    line(blockRightX, blockBottomY, netColX, blockBottomY);
                    // netColX vertical: from blockBottomY up to tipY
                    line(netColX, blockBottomY, netColX, tipY);
                    trackX(daColX - 5);
                } else {
                    // DA bare: horizontal at tipY already at netColX level
                    line(daColX, tipY, netColX, tipY);
                    trackX(daColX - 5);
                }

                // NET (normal, RIGHT) bare: horizontal at tipY → right to returnRightX
                // (netColX is already connected to tipY from DA path above)
                line(netColX, tipY, returnRightX, tipY);
                // returnRightX → up → arrow to while start
                line(returnRightX, tipY, returnRightX, contTargetY);
                arrow(returnRightX, contTargetY, x + 5, contTargetY);
                trackX(netColX + 5);
                trackX(returnRightX + 5);

            } else {
                // ── BREAK DECISION routing ─────────────────────────────────
                double breakTargetY = y - VERTICAL_SPACING / 2;

                if (netBare) {
                    line(netColX, tipY, returnRightX, tipY);
                    line(returnRightX, tipY, returnRightX, breakTargetY);
                    arrow(returnRightX, breakTargetY, x + 5, breakTargetY);
                    trackX(returnRightX + 5);
                } else {
                    double netCornerY = netLineStartY + VERTICAL_SPACING / 2;
                    line(netColX, netLineStartY, netColX, netCornerY);
                    line(netColX, netCornerY, returnRightX, netCornerY);
                    line(returnRightX, netCornerY, returnRightX, breakTargetY);
                    arrow(returnRightX, breakTargetY, x + 5, breakTargetY);
                    trackX(returnRightX + 5);
                }

                double breakJoinY = exitY - VERTICAL_SPACING / 2;
                if (daBare) {
                    line(daColX, tipY, daColX, breakJoinY);
                } else {
                    line(daColX, daColEndY, daColX, breakJoinY);
                }
                arrow(daColX, breakJoinY, x - 1, breakJoinY);
                trackX(daColX - 5);
            }
        }

        if (breakGeometry.isEmpty()) {
            double backStartX  = Double.isNaN(lastBodyBlockX) ? rightX : lastBodyBlockX;
            double backCornerY = bodyEndY + VERTICAL_SPACING / 2;
            double backTargetY = y - VERTICAL_SPACING / 2;
            line(backStartX, bodyEndY, backStartX, backCornerY);
            line(backStartX, backCornerY, returnRightX, backCornerY);
            line(returnRightX, backCornerY, returnRightX, backTargetY);
            arrow(returnRightX, backTargetY, x + 5, backTargetY);
            trackX(returnRightX + 5);
        }

        updateMaxY(bodyEndY);

        if (node.getExitNode() != null) {
            arrow(x, diamondBottom, x, exitY - 5);

            rendered.add(node.getExitNode());
            node.getExitNode().setPosition(x, exitY);
            renderProcess(node.getExitNode(), x, exitY);
            node.getExitNode().setSize(PROCESS_WIDTH, PROCESS_HEIGHT);
            updateMaxY(exitY + PROCESS_HEIGHT);

            double exitBottom = exitY + PROCESS_HEIGHT;
            double afterExitY = exitBottom + VERTICAL_SPACING;

            List<FlowchartNode> exitNextList = node.getExitNode().getNext();
            if (!exitNextList.isEmpty()) {
                FlowchartNode exitNext = exitNextList.get(0);
                if (exitNext instanceof TerminalNode t && !t.isStart()) {
                    endNode = t;
                    endArrowFromX = x;
                    endArrowFromY = exitBottom;
                } else if (!(exitNext instanceof LoopEndNode) && !rendered.contains(exitNext)) {
                    arrow(x, exitBottom, x, afterExitY - 5);
                    renderNode(exitNext, x, afterExitY, stopBefore);
                }
            }

            node.setSize(w, exitBottom - y);
        } else {
            line(x, diamondBottom, x, exitY);
            updateMaxY(exitY);
            node.setSize(w, exitY - y);
        }
    }

    private List<FlowchartNode> collectBodyChain(FlowchartNode start, LoopStartNode loop) {
        List<FlowchartNode> chain = new ArrayList<>();
        IdentityHashMap<FlowchartNode, Boolean> seen = new IdentityHashMap<>();
        FlowchartNode cur = start;

        while (cur != null && !seen.containsKey(cur)) {
            if (cur instanceof LoopEndNode)  break;
            if (cur instanceof TerminalNode) break;
            if (cur == loop.getExitNode())   break;

            seen.put(cur, true);
            chain.add(cur);

            FlowchartNode next = null;
            for (FlowchartNode n : cur.getNext()) {
                if (n instanceof LoopEndNode)       continue;
                if (n instanceof TerminalNode)      continue;
                if (n == loop.getExitNode())        continue;
                if (cur instanceof DecisionNode dn) {
                    if (n == dn.getTrueBranch())  continue;
                    if (n == dn.getFalseBranch()) continue;
                }
                next = n;
                break;
            }
            cur = next;
        }
        return chain;
    }

    private double renderLoopBodyChain(List<FlowchartNode> chain, double x, double startY, LoopStartNode loop) {
        double currentY = startY;
        final double bodyStartX = x;
        final double bodyStartY = startY;

        rendered.addAll(chain);

        for (int i = 0; i < chain.size(); i++) {
            FlowchartNode node     = chain.get(i);
            FlowchartNode nextNode = (i + 1 < chain.size()) ? chain.get(i + 1) : null;

            if (node instanceof ProcessNode) {
                rendered.add(node);
                node.setPosition(x, currentY);
                renderProcess(node, x, currentY);
                trackX(x + PROCESS_WIDTH / 2);
                updateMaxY(currentY + PROCESS_HEIGHT);

                double nodeBottom = currentY + PROCESS_HEIGHT;
                if (nextNode != null) {
                    double nextY = nodeBottom + VERTICAL_SPACING;
                    arrow(x, nodeBottom, x, nextY - 5);
                    currentY = nextY;
                } else {
                    lastBodyBlockX = x;
                    return nodeBottom;
                }

            } else if (node instanceof DecisionNode decisionNode) {
                rendered.add(node);
                node.setPosition(x, currentY);

                double mergeY = renderDecisionInBody(decisionNode, x, currentY, nextNode, loop, bodyStartX, bodyStartY, null, null);

                if (mergeY < 0) {
                    double blockBottom = -mergeY;

                    if (nextNode != null) {
                        double[] g = breakGeometry.get(decisionNode);
                        if (g != null) {
                            double tailColX      = g[1];
                            double colStartY     = g[6];
                            boolean columnWasBare = (g[3] <= colStartY + 1);
                            double tailStartY    = columnWasBare ? colStartY : g[3];

                            if (columnWasBare) {
                                double netTipY = g[5];
                                arrow(tailColX, netTipY, tailColX, colStartY);
                            }

                            List<FlowchartNode> tail = chain.subList(i + 1, chain.size());
                            rendered.addAll(tail);
                            double tailEndY = renderBreakTailChain(tail, tailColX, tailStartY, columnWasBare);
                            g[3] = tailEndY;
                            blockBottom = Math.max(-mergeY, tailEndY);
                        }
                    }
                    return -blockBottom;
                }

                if (nextNode != null) {
                    double nextY = mergeY + VERTICAL_SPACING;
                    arrow(x, mergeY, x, nextY - 5);
                    currentY = nextY;
                } else {
                    double lineEndY = mergeY + VERTICAL_SPACING;
                    line(x, mergeY, x, lineEndY);
                    updateMaxY(lineEndY);
                    return lineEndY;
                }

            } else {
                break;
            }
        }

        return currentY;
    }

    private double renderDecisionInBody(DecisionNode node, double x, double y,
                                        FlowchartNode stopBefore, LoopStartNode loop,
                                        double loopBodyX, double loopBodyY,
                                        DoWhileNode doWhileLoop,
                                        List<double[]> continueExits) {
        double w = DECISION_WIDTH, h = DECISION_HEIGHT;
        double halfW = w / 2, halfH = h / 2;
        node.setSize(w, h);
        drawDiamond(x, y, w, h);
        text(node.getLabel(), x, y + halfH);

        boolean trueEndsWithBreak    = chainEndsWithBreak(node.getTrueBranch());
        boolean falseEndsWithBreak   = chainEndsWithBreak(node.getFalseBranch());
        boolean trueEndsWithContinue = chainEndsWithContinue(node.getTrueBranch());
        boolean falseEndsWithContinue= chainEndsWithContinue(node.getFalseBranch());
        boolean isBreakDecision      = trueEndsWithBreak  || falseEndsWithBreak;
        boolean isContinueDecision   = trueEndsWithContinue || falseEndsWithContinue;

        boolean insideDoWhile = (doWhileLoop != null);

        // ── CONTINUE DECISION ─────────────────────────────────────────────────
        if (isContinueDecision && !isBreakDecision) {
            FlowchartNode contBranch   = trueEndsWithContinue ? node.getTrueBranch()  : node.getFalseBranch();
            FlowchartNode normalBranch = trueEndsWithContinue ? node.getFalseBranch() : node.getTrueBranch();
            String contLabel   = trueEndsWithContinue ? "ДА"  : "НЕТ";
            String normalLabel = trueEndsWithContinue ? "НЕТ" : "ДА";

            if (insideDoWhile) {
                double rightTipX    = x + halfW;
                double rightTipY    = y + halfH;
                double normalStartY = y + h + VERTICAL_SPACING;
                double bottomTipX   = x;
                double bottomTipY   = y + h;

                labelText(contLabel, rightTipX + 10, rightTipY - 10);
                deferredDoWhileContinueCols.add(new Object[]{rightTipX, rightTipY, contBranch, continueExits});
                labelText(normalLabel, bottomTipX + 8, bottomTipY + 15);

                double normalEndY;
                if (normalBranch != null) {
                    normalEndY = renderInlineChain(normalBranch, doWhileLoop, bottomTipX, normalStartY, continueExits, bottomTipY);
                } else {
                    normalEndY = bottomTipY;
                }

                node.setSize(w, normalEndY - y);
                updateMaxY(normalEndY);
                return normalEndY;

            } else {
                // ── While/for continue decision ──────────────────────────────
                double colOffset = BREAK_HORIZONTAL_SPACING;
                double leftColX  = x - colOffset;
                double rightColX = x + colOffset;
                trackX(leftColX  - PROCESS_WIDTH / 2);
                trackX(rightColX + PROCESS_WIDTH / 2);

                double tipY      = y + halfH;
                double colStartY = y + h + VERTICAL_SPACING;

                // ── ДА (continue) column ─────────────────────────────────────
                labelText(contLabel, x - halfW - 30, tipY - 10);
                line(x - halfW, tipY, leftColX, tipY);

                double leftColEndY;
                if (isContinue(contBranch)) {
                    leftColEndY = colStartY;
                } else {
                    arrow(leftColX, tipY, leftColX, colStartY);
                    leftColEndY = renderBreakColumnChain(contBranch, leftColX, colStartY);
                }

                // ── НЕТ (normal) column ──────────────────────────────────────
                labelText(normalLabel, x + halfW + 10, tipY - 10);
                line(x + halfW, tipY, rightColX, tipY);

                double rightColEndY;
                if (normalBranch != null && !isContinue(normalBranch) && !isBreak(normalBranch)) {
                    arrow(rightColX, tipY, rightColX, colStartY);
                    rightColEndY = renderBreakColumnChain(normalBranch, rightColX, colStartY);
                } else {
                    rightColEndY = colStartY;
                }

                breakGeometry.put(node, new double[]{leftColX, rightColX, leftColEndY, rightColEndY, 1.0, tipY, colStartY});

                double blockBottom = Math.max(leftColEndY, rightColEndY) + VERTICAL_SPACING;
                node.setSize(w, blockBottom - y);
                updateMaxY(blockBottom);
                return -blockBottom;
            }
        }

        // ── BREAK DECISION ────────────────────────────────────────────────────
        if (isBreakDecision) {
            FlowchartNode breakBranch    = trueEndsWithBreak ? node.getTrueBranch()  : node.getFalseBranch();
            FlowchartNode continueBranch = trueEndsWithBreak ? node.getFalseBranch() : node.getTrueBranch();
            String breakLabel    = trueEndsWithBreak ? "ДА"  : "НЕТ";
            String continueLabel = trueEndsWithBreak ? "НЕТ" : "ДА";

            double colOffset = BREAK_HORIZONTAL_SPACING;
            double daColX    = x - colOffset;
            double netColX   = x + colOffset;
            trackX(daColX  - PROCESS_WIDTH / 2);
            trackX(netColX + PROCESS_WIDTH / 2);

            double tipY      = y + halfH;
            double colStartY = y + h + VERTICAL_SPACING;

            // ── ДА (break) column ────────────────────────────────────────────
            labelText(breakLabel, x - halfW - 30, tipY - 10);
            line(x - halfW, tipY, daColX, tipY);

            double daColEndY;
            if (isBreak(breakBranch)) {
                daColEndY = colStartY;
            } else {
                arrow(daColX, tipY, daColX, colStartY);
                daColEndY = renderBreakColumnChain(breakBranch, daColX, colStartY);
            }

            // ── НЕТ (continue-in-loop) column ───────────────────────────────
            labelText(continueLabel, x + halfW + 10, tipY - 10);
            line(x + halfW, tipY, netColX, tipY);

            double netColEndY;
            boolean netIsBare = (continueBranch == null || isContinue(continueBranch) || isBreak(continueBranch));
            if (!netIsBare) {
                arrow(netColX, tipY, netColX, colStartY);
                netColEndY = renderBreakColumnChain(continueBranch, netColX, colStartY);
            } else {
                netColEndY = colStartY;
            }

            double netBareFlag = netIsBare ? 1.0 : 0.0;
            breakGeometry.put(node, new double[]{daColX, netColX, daColEndY, netColEndY, 0.0, tipY, colStartY, netBareFlag});

            double blockBottom = Math.max(daColEndY, netColEndY) + VERTICAL_SPACING;
            node.setSize(w, blockBottom - y);
            updateMaxY(blockBottom);
            return -blockBottom;
        }

        // ── NORMAL DECISION ───────────────────────────────────────────────────
        double branchY = y + h + VERTICAL_SPACING;
        double leftX   = x - HORIZONTAL_SPACING;
        double rightX  = x + HORIZONTAL_SPACING;
        trackX(leftX - PROCESS_WIDTH / 2);
        trackX(rightX + PROCESS_WIDTH / 2);

        double leftBottom  = branchY;
        double rightBottom = branchY;

        if (node.getTrueBranch() != null) {
            line(x - halfW, y + halfH, leftX, y + halfH);
            line(leftX, y + halfH, leftX, branchY - 5);
            arrow(leftX, branchY - 5, leftX, branchY);
            labelText("ДА", x - halfW - 30, y + halfH - 10);
            renderNode(node.getTrueBranch(), leftX, branchY, stopBefore);
            leftBottom = branchY + branchHeight(node.getTrueBranch(), stopBefore);
        }

        if (node.getFalseBranch() != null) {
            line(x + halfW, y + halfH, rightX, y + halfH);
            line(rightX, y + halfH, rightX, branchY - 5);
            arrow(rightX, branchY - 5, rightX, branchY);
            labelText("НЕТ", x + halfW + 10, y + halfH - 10);
            renderNode(node.getFalseBranch(), rightX, branchY, stopBefore);
            rightBottom = branchY + branchHeight(node.getFalseBranch(), stopBefore);
        } else {
            rightBottom = leftBottom;
        }

        double mergeY = Math.max(leftBottom, rightBottom) + VERTICAL_SPACING;

        if (node.getTrueBranch() != null) {
            line(leftX, leftBottom, leftX, mergeY);
            line(leftX, mergeY, x, mergeY);
        }
        if (node.getFalseBranch() != null) {
            line(rightX, rightBottom, rightX, mergeY);
            line(rightX, mergeY, x, mergeY);
        } else {
            double noElseRightX = x + HORIZONTAL_SPACING;
            line(x + halfW, y + halfH, noElseRightX, y + halfH);
            labelText("НЕТ", x + halfW + 10, y + halfH - 10);
            line(noElseRightX, y + halfH, noElseRightX, mergeY);
            line(noElseRightX, mergeY, x, mergeY);
            trackX(noElseRightX + 5);
        }

        node.setSize(w, mergeY - y);
        updateMaxY(mergeY);
        return mergeY;
    }

    private double renderBreakColumnChain(FlowchartNode start, double x, double startY) {
        double currentY = startY;
        FlowchartNode cur = start;

        while (cur != null) {
            if (cur instanceof ConnectorNode) break;
            if (cur instanceof LoopEndNode)   break;
            if (cur instanceof TerminalNode)  break;

            if (cur instanceof ProcessNode) {
                rendered.add(cur);
                cur.setPosition(x, currentY);
                renderProcess(cur, x, currentY);
                trackX(x + PROCESS_WIDTH / 2);
                updateMaxY(currentY + PROCESS_HEIGHT);

                double nodeBottom = currentY + PROCESS_HEIGHT;

                FlowchartNode next = null;
                for (FlowchartNode n : cur.getNext()) {
                    if (n instanceof ConnectorNode) break;
                    if (n instanceof LoopEndNode)   break;
                    next = n;
                    break;
                }

                if (next != null) {
                    double nextY = nodeBottom + VERTICAL_SPACING;
                    arrow(x, nodeBottom, x, nextY - 5);
                    currentY = nextY;
                    cur = next;
                } else {
                    return nodeBottom;
                }
            } else {
                break;
            }
        }

        return currentY;
    }

    private double renderBreakTailChain(List<FlowchartNode> tail, double x, double startY, boolean columnWasBare) {
        double currentY = startY;

        for (int i = 0; i < tail.size(); i++) {
            FlowchartNode node = tail.get(i);

            if (node instanceof ProcessNode) {
                rendered.add(node);

                double blockY;
                if (i == 0 && columnWasBare) {
                    blockY = currentY;
                } else {
                    blockY = currentY + VERTICAL_SPACING;
                    arrow(x, currentY, x, blockY - 5);
                }
                currentY = blockY;

                node.setPosition(x, currentY);
                renderProcess(node, x, currentY);
                trackX(x + PROCESS_WIDTH / 2);
                updateMaxY(currentY + PROCESS_HEIGHT);

                double nodeBottom = currentY + PROCESS_HEIGHT;
                if (i + 1 < tail.size()) {
                    double nextY = nodeBottom + VERTICAL_SPACING;
                    arrow(x, nodeBottom, x, nextY - 5);
                    currentY = nextY;
                } else {
                    lastBodyBlockX = x;
                    return nodeBottom;
                }
            } else {
                break;
            }
        }

        return currentY;
    }

    private double renderInlineChain(FlowchartNode start, DoWhileNode loop,
                                     double x, double startY, List<double[]> continueExits,
                                     double fromY) {
        if (fromY >= 0) {
            line(x, fromY, x, startY - 5);
            arrow(x, startY - 5, x, startY);
        }

        double currentY = startY;
        FlowchartNode cur = start;
        IdentityHashMap<FlowchartNode, Boolean> seen = new IdentityHashMap<>();

        while (cur != null && !seen.containsKey(cur)) {
            if (cur instanceof LoopEndNode)  break;
            if (cur instanceof TerminalNode) break;
            if (cur == loop.getExitNode())   break;

            seen.put(cur, true);
            rendered.add(cur);
            cur.setPosition(x, currentY);

            if (cur instanceof ProcessNode) {
                renderProcess(cur, x, currentY);
                trackX(x + PROCESS_WIDTH / 2);
                updateMaxY(currentY + PROCESS_HEIGHT);
                double nodeBottom = currentY + PROCESS_HEIGHT;

                FlowchartNode next = null;
                for (FlowchartNode n : cur.getNext()) {
                    if (n instanceof LoopEndNode)  continue;
                    if (n instanceof TerminalNode) continue;
                    if (n == loop.getExitNode())   continue;
                    next = n;
                    break;
                }

                if (next != null) {
                    if (next instanceof ConnectorNode conn && "continue".equals(conn.getLabel())) {
                        double rightEdgeX = x + PROCESS_WIDTH / 2;
                        double midBlockY  = currentY + PROCESS_HEIGHT / 2;
                        continueExits.add(new double[]{rightEdgeX, midBlockY});
                        rendered.add(next);
                        return -nodeBottom;
                    }
                    double nextY = nodeBottom + VERTICAL_SPACING;
                    arrow(x, nodeBottom, x, nextY - 5);
                    currentY = nextY;
                    cur = next;
                } else {
                    return nodeBottom;
                }

            } else {
                break;
            }
        }

        return currentY;
    }

    private boolean chainEndsWithBreak(FlowchartNode node) {
        if (node == null) return false;
        if (isBreak(node)) return true;
        if (node instanceof ProcessNode) {
            for (FlowchartNode next : node.getNext()) {
                if (chainEndsWithBreak(next)) return true;
            }
        }
        return false;
    }

    private boolean isBreak(FlowchartNode node) {
        return node instanceof ConnectorNode c && "break".equals(c.getLabel());
    }

    private boolean isContinue(FlowchartNode node) {
        return node instanceof ConnectorNode c && "continue".equals(c.getLabel());
    }

    private boolean chainEndsWithContinue(FlowchartNode node) {
        if (node == null) return false;
        if (isContinue(node)) return true;
        if (node instanceof ProcessNode) {
            for (FlowchartNode next : node.getNext()) {
                if (chainEndsWithContinue(next)) return true;
            }
        }
        return false;
    }

    private void renderLinearNext(FlowchartNode node, double x, double y, FlowchartNode stopBefore) {
        if (node.getNext().isEmpty()) return;
        double prevBottom = y + node.getHeight();

        for (FlowchartNode next : node.getNext()) {
            if (next == stopBefore) continue;

            if (next instanceof TerminalNode t && !t.isStart()) {
                endNode = t;
                endArrowFromX = x;
                endArrowFromY = prevBottom;
                return;
            }

            double nextY = prevBottom + VERTICAL_SPACING;
            arrow(x, prevBottom, x, nextY - 5);
            renderNode(next, x, nextY, stopBefore);
        }
    }

    // ── DO-WHILE ──────────────────────────────────────────────────────────────
    private void renderDoWhile(DoWhileNode node, double x, double y, FlowchartNode stopBefore) {
        double arrowMidY = y - VERTICAL_SPACING / 2;

        List<double[]> continueExits = new ArrayList<>();

        double bodyEndY = renderDoWhileBody(node.getLoopBody(), node, x, y, continueExits);
        if (bodyEndY < 0) bodyEndY = -bodyEndY;

        for (Object[] deferred : deferredDoWhileContinueCols) {
            double rightTipX2  = (double) deferred[0];
            double rightTipY2  = (double) deferred[1];
            FlowchartNode contBr = (FlowchartNode) deferred[2];
            @SuppressWarnings("unchecked")
            List<double[]> exits = (List<double[]>) deferred[3];

            boolean bareContinue = isContinue(contBr);

            if (bareContinue) {
                exits.add(new double[]{rightTipX2, rightTipY2, 1.0, 1.0});
            } else {
                double contColX2  = maxX + BACK_ARROW_MARGIN + PROCESS_WIDTH / 2 + 20;
                double colStartY2 = rightTipY2 + VERTICAL_SPACING;
                trackX(contColX2 + PROCESS_WIDTH / 2);
                line(rightTipX2, rightTipY2, contColX2, rightTipY2);
                line(contColX2, rightTipY2, contColX2, colStartY2);
                double contColEndY2 = renderBreakColumnChain(contBr, contColX2, colStartY2);
                exits.add(new double[]{contColX2, contColEndY2, 1.0});
                bodyEndY = Math.max(bodyEndY, contColEndY2);
            }
        }
        deferredDoWhileContinueCols.clear();

        double dY = bodyEndY + VERTICAL_SPACING;
        arrow(x, bodyEndY, x, dY - 5);

        double dW = DECISION_WIDTH, dH = DECISION_HEIGHT;
        double halfW = dW / 2, halfH = dH / 2;
        node.setPosition(x, dY);
        drawDiamond(x, dY, dW, dH);
        text(node.getLabel(), x, dY + halfH);
        updateMaxY(dY + dH);

        boolean hasContinue = !continueExits.isEmpty();
        double rightTipX = x + halfW;
        double rightTipY = dY + halfH;

        if (hasContinue) {
            double contColX = continueExits.stream().mapToDouble(e -> e[0]).max()
                    .orElse(rightTipX + BREAK_HORIZONTAL_SPACING);
            contColX = Math.max(contColX, x + halfW + BACK_ARROW_MARGIN);
            trackX(contColX + 5);

            for (double[] exit : continueExits) {
                line(exit[0], exit[1], contColX, exit[1]);
            }

            double topY = continueExits.stream().mapToDouble(e -> e[1]).min().orElse(rightTipY);
            line(contColX, topY, contColX, rightTipY);
            arrow(contColX, rightTipY, rightTipX + 1, rightTipY);

            boolean anyNeedsLabel = continueExits.stream().anyMatch(e -> e.length < 3 || e[2] != 1.0);
            if (anyNeedsLabel) labelText("ДА", contColX - 35, topY - 10);
        }

        double leftColX = minX - BACK_ARROW_MARGIN;
        double tipY = dY + halfH;
        double leftTipX = x - halfW;
        line(leftTipX, tipY, leftColX, tipY);
        line(leftColX, tipY, leftColX, arrowMidY);
        arrow(leftColX, arrowMidY, x - 1, arrowMidY);
        labelText("ДА", leftTipX - 40, tipY - 10);
        trackX(leftColX - 5);

        double dBottom = dY + dH;
        labelText("НЕТ", x + 8, dBottom + 15);

        if (node.getExitNode() != null) {
            FlowchartNode exitNode = node.getExitNode();
            if (exitNode instanceof TerminalNode t && !t.isStart()) {
                endNode = t;
                endArrowFromX = x;
                endArrowFromY = dBottom;
                node.setSize(dW, dBottom - y);
            } else {
                double exitY = dBottom + VERTICAL_SPACING;
                arrow(x, dBottom, x, exitY - 5);
                rendered.add(exitNode);
                exitNode.setPosition(x, exitY);
                renderProcess(exitNode, x, exitY);
                exitNode.setSize(PROCESS_WIDTH, PROCESS_HEIGHT);
                updateMaxY(exitY + PROCESS_HEIGHT);
                double exitBottom = exitY + PROCESS_HEIGHT;
                List<FlowchartNode> exitNextList = exitNode.getNext();
                if (!exitNextList.isEmpty()) {
                    FlowchartNode exitNext = exitNextList.get(0);
                    if (exitNext instanceof TerminalNode t && !t.isStart()) {
                        endNode = t;
                        endArrowFromX = x;
                        endArrowFromY = exitBottom;
                    } else if (!(exitNext instanceof LoopEndNode) && !rendered.contains(exitNext)) {
                        double afterExitY = exitBottom + VERTICAL_SPACING;
                        arrow(x, exitBottom, x, afterExitY - 5);
                        renderNode(exitNext, x, afterExitY, stopBefore);
                    }
                }
                node.setSize(dW, exitBottom - y);
            }
        } else {
            endArrowFromX = x;
            endArrowFromY = dBottom;
            node.setSize(dW, dBottom - y);
        }
    }

    private double renderDoWhileBody(FlowchartNode start, DoWhileNode loop, double x, double startY,
                                     List<double[]> continueExits) {
        double currentY = startY;
        FlowchartNode cur = start;
        IdentityHashMap<FlowchartNode, Boolean> seen = new IdentityHashMap<>();

        while (cur != null && !seen.containsKey(cur)) {
            if (cur instanceof LoopEndNode)   break;
            if (cur instanceof TerminalNode)  break;
            if (cur == loop.getExitNode())    break;

            seen.put(cur, true);

            if (cur instanceof ConnectorNode conn && "continue".equals(conn.getLabel())) {
                continueExits.add(new double[]{x + PROCESS_WIDTH / 2, currentY});
                rendered.add(cur);
                return -currentY;
            }

            rendered.add(cur);
            cur.setPosition(x, currentY);

            if (cur instanceof ProcessNode) {
                renderProcess(cur, x, currentY);
                updateMaxY(currentY + PROCESS_HEIGHT);
                trackX(x + PROCESS_WIDTH / 2);
                double nodeBottom = currentY + PROCESS_HEIGHT;

                FlowchartNode next = null;
                for (FlowchartNode n : cur.getNext()) {
                    if (n instanceof LoopEndNode)  continue;
                    if (n instanceof TerminalNode) continue;
                    if (n == loop.getExitNode())   continue;
                    next = n;
                    break;
                }

                if (next != null) {
                    if (next instanceof ConnectorNode conn && "continue".equals(conn.getLabel())) {
                        double rightEdgeX = x + PROCESS_WIDTH / 2;
                        double midBlockY  = currentY + PROCESS_HEIGHT / 2;
                        continueExits.add(new double[]{rightEdgeX, midBlockY});
                        rendered.add(next);
                        return -nodeBottom;
                    }
                    double nextY = nodeBottom + VERTICAL_SPACING;
                    arrow(x, nodeBottom, x, nextY - 5);
                    currentY = nextY;
                    cur = next;
                } else {
                    return nodeBottom;
                }

            } else if (cur instanceof DecisionNode decNode) {
                boolean trueIsContinue  = chainEndsWithContinue(decNode.getTrueBranch());
                boolean falseIsContinue = chainEndsWithContinue(decNode.getFalseBranch());
                boolean hasContinueBranch = trueIsContinue || falseIsContinue;

                FlowchartNode afterIf = null;
                for (FlowchartNode n : decNode.getNext()) {
                    if (n instanceof LoopEndNode)      continue;
                    if (n instanceof TerminalNode)     continue;
                    if (n == loop.getExitNode())       continue;
                    if (n == decNode.getTrueBranch())  continue;
                    if (n == decNode.getFalseBranch()) continue;
                    afterIf = n;
                    break;
                }

                double decBottom;
                if (hasContinueBranch) {
                    double mergeY = renderDecisionInBody(decNode, x, currentY, afterIf, null,
                            x, currentY, loop, continueExits);
                    decBottom = mergeY;
                } else {
                    renderDecision(decNode, x, currentY, afterIf);
                    decBottom = currentY + decNode.getHeight();
                }

                if (afterIf != null && !rendered.contains(afterIf)) {
                    double nextY = decBottom + VERTICAL_SPACING;
                    arrow(x, decBottom, x, nextY - 5);
                    currentY = nextY;
                    cur = afterIf;
                } else {
                    return decBottom;
                }

            } else {
                break;
            }
        }

        return currentY;
    }

    private TerminalNode findEndTerminal(FlowchartNode node) {
        return findEndTerminalRecursive(node, new HashSet<>());
    }

    private TerminalNode findEndTerminalRecursive(FlowchartNode node, Set<FlowchartNode> visited) {
        if (node == null || visited.contains(node)) return null;
        visited.add(node);
        if (node instanceof TerminalNode t && !t.isStart()) return t;
        for (FlowchartNode next : node.getNext()) {
            TerminalNode found = findEndTerminalRecursive(next, visited);
            if (found != null) return found;
        }
        if (node instanceof DecisionNode d) {
            TerminalNode found = findEndTerminalRecursive(d.getTrueBranch(), visited);
            if (found != null) return found;
            found = findEndTerminalRecursive(d.getFalseBranch(), visited);
            if (found != null) return found;
        }
        return null;
    }

    private void drawDiamond(double x, double y, double w, double h) {
        double halfW = w / 2, halfH = h / 2;
        String points = String.format(Locale.US,
                "%.1f,%.1f %.1f,%.1f %.1f,%.1f %.1f,%.1f",
                x, y, x + halfW, y + halfH, x, y + h, x - halfW, y + halfH);
        svg.append(String.format(Locale.US, "<polygon class='shape' points='%s'/>\n", points));
    }

    private void arrow(double x1, double y1, double x2, double y2) {
        svg.append(String.format(Locale.US,
                "<line class=\"arrow\" x1=\"%.1f\" y1=\"%.1f\" x2=\"%.1f\" y2=\"%.1f\"/>\n",
                x1, y1, x2, y2));
    }

    private void line(double x1, double y1, double x2, double y2) {
        svg.append(String.format(Locale.US,
                "<line class=\"line\" x1=\"%.1f\" y1=\"%.1f\" x2=\"%.1f\" y2=\"%.1f\"/>\n",
                x1, y1, x2, y2));
    }

    private void text(String txt, double x, double y) {
        svg.append(String.format(Locale.US,
                "<text class=\"text\" x=\"%.1f\" y=\"%.1f\">%s</text>\n",
                x, y, escapeXml(txt)));
    }

    private void labelText(String txt, double x, double y) {
        svg.append(String.format(Locale.US,
                "<text class=\"label\" x=\"%.1f\" y=\"%.1f\">%s</text>\n",
                x, y, escapeXml(txt)));
    }

    private String escapeXml(String s) {
        return s.replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;");
    }
}