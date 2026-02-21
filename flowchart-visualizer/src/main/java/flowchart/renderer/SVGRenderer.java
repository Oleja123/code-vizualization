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

    // Extra margin to the right of all body blocks for the back-arrow
    private static final double BACK_ARROW_MARGIN = 40;

    private StringBuilder svg;
    private Set<FlowchartNode> rendered;
    private Map<DecisionNode, double[]> breakGeometry; // daColX, netColX, daColEndY, netColEndY
    private double maxY = 0;
    private double minX = Double.MAX_VALUE;
    private double maxX = Double.MIN_VALUE;

    private TerminalNode endNode;
    private double endArrowFromX;
    private double endArrowFromY;

    // Текущий returnRightX активного цикла
    private double currentLoopReturnRightX = Double.MAX_VALUE;

    // X-координата последнего блока тела цикла (для правильного старта back-arrow)
    private double lastBodyBlockX = Double.NaN;

    public String render(FlowchartNode start) {
        svg      = new StringBuilder();
        rendered = new HashSet<>();
        breakGeometry = new HashMap<>();
        maxY = 0;
        minX = Double.MAX_VALUE;
        maxX = Double.MIN_VALUE;
        endNode = null;
        endArrowFromX = 700;
        endArrowFromY = 0;
        currentLoopReturnRightX = Double.MAX_VALUE;
        lastBodyBlockX = Double.NaN;

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
    //
    // KEY CHANGES vs original:
    // 1. returnRightX is computed AFTER rendering the body (using actual maxX + margin),
    //    so the back-arrow always clears all blocks including the last body node (i++).
    // 2. The back-arrow horizontal line starts from the bodyEndY returned by renderLoopBodyChain,
    //    which is the real bottom of the last block (not an artificial line-end Y).
    // 3. break-columns use a fixed BREAK_HORIZONTAL_SPACING without clamping to a
    //    pre-render returnRightX that was too small.

    private void renderLoop(LoopStartNode node, double x, double y, FlowchartNode stopBefore) {
        double w = DECISION_WIDTH, h = DECISION_HEIGHT;
        double halfW = w / 2, halfH = h / 2;

        drawDiamond(x, y, w, h);
        text(node.getLabel(), x, y + halfH);

        double rightX  = x + HORIZONTAL_SPACING;
        double branchY = y + h + VERTICAL_SPACING;

        // Arrow ДА: from right tip of diamond → right → down into body
        line(x + halfW, y + halfH, rightX, y + halfH);
        arrow(rightX, y + halfH, rightX, branchY - 5);
        labelText("ДА", x + halfW + 10, y + halfH - 10);

        // Set a large placeholder so break-column offset calculation in renderDecisionInBody
        // doesn't clip prematurely. We'll use maxX+margin after body is rendered.
        currentLoopReturnRightX = Double.MAX_VALUE;

        // Render body chain
        List<FlowchartNode> bodyChain = collectBodyChain(node.getLoopBody(), node);
        double bodyEndY = renderLoopBodyChain(bodyChain, rightX, branchY, node);
        if (bodyEndY < 0) bodyEndY = -bodyEndY; // unwrap sentinel

        // --- COMPUTE returnRightX AFTER body render ---
        double returnRightX = maxX + BACK_ARROW_MARGIN;
        currentLoopReturnRightX = returnRightX;

        // НЕТ exits straight down from diamond bottom.
        double diamondBottom = y + h;
        labelText("НЕТ", x + 8, diamondBottom + 15);

        // maxCornerY: the lowest Y where continue-paths turn right before going up.
        // exitY for the loop НЕТ branch must be below this.
        double maxCornerY = bodyEndY + VERTICAL_SPACING / 2;

        for (Map.Entry<DecisionNode, double[]> entry : breakGeometry.entrySet()) {
            double[] g = entry.getValue();
            double daColX     = g[0];
            double netColX    = g[1];
            double daColEndY  = g[2];
            double netColEndY = g[3];

            // Continue column: short drop from bottom of last block (i++) → right → up to diamond
            double netCornerY = netColEndY + VERTICAL_SPACING / 2;
            maxCornerY = Math.max(maxCornerY, netCornerY);

            line(netColX, netColEndY, netColX, netCornerY);       // ↓ short drop from i++
            line(netColX, netCornerY, returnRightX, netCornerY);  // → right to margin
            line(returnRightX, netCornerY, returnRightX, y);      // ↑ up to diamond
            arrow(returnRightX, y, x + 5, y);                     // ← arrow to diamond top
            trackX(returnRightX + 5);
        }

        double exitY = maxCornerY + VERTICAL_SPACING / 2;

        // Break columns: vertical down to just above exitY → horizontal arrow to loop axis
        for (Map.Entry<DecisionNode, double[]> entry : breakGeometry.entrySet()) {
            double[] g = entry.getValue();
            double daColX    = g[0];
            double daColEndY = g[2];

            double breakJoinY = exitY - VERTICAL_SPACING / 2;
            line(daColX, daColEndY, daColX, breakJoinY);
            arrow(daColX, breakJoinY, x - 1, breakJoinY);
            trackX(daColX - 5);
        }

        // If there was no break-decision, draw the normal back-arrow from lastBodyBlockX
        if (breakGeometry.isEmpty()) {
            double backStartX  = Double.isNaN(lastBodyBlockX) ? rightX : lastBodyBlockX;
            double backCornerY = bodyEndY + VERTICAL_SPACING / 2;
            line(backStartX, bodyEndY, backStartX, backCornerY);       // ↓ short drop
            line(backStartX, backCornerY, returnRightX, backCornerY);  // → right
            line(returnRightX, backCornerY, returnRightX, y);          // ↑ up
            arrow(returnRightX, y, x + 5, y);                          // ← arrow to diamond
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
                    // Last block in body — save its X so back-arrow starts from correct column
                    lastBodyBlockX = x;
                    return nodeBottom;
                }

            } else if (node instanceof DecisionNode decisionNode) {
                rendered.add(node);
                node.setPosition(x, currentY);

                double mergeY = renderDecisionInBody(decisionNode, x, currentY, nextNode, loop);

                if (mergeY < 0) {
                    double blockBottom = -mergeY;

                    if (nextNode != null) {
                        double[] g = breakGeometry.get(decisionNode);
                        if (g != null) {
                            double netColX   = g[1];
                            double netStartY = g[3];
                            List<FlowchartNode> tail = chain.subList(i + 1, chain.size());
                            rendered.addAll(tail);
                            double tailEndY = renderBreakTailChain(tail, netColX, netStartY);
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
                                        FlowchartNode stopBefore, LoopStartNode loop) {
        double w = DECISION_WIDTH, h = DECISION_HEIGHT;
        double halfW = w / 2, halfH = h / 2;
        node.setSize(w, h);
        drawDiamond(x, y, w, h);
        text(node.getLabel(), x, y + halfH);

        boolean trueEndsWithBreak  = chainEndsWithBreak(node.getTrueBranch());
        boolean falseEndsWithBreak = chainEndsWithBreak(node.getFalseBranch());
        boolean isBreakDecision    = trueEndsWithBreak || falseEndsWithBreak;

        if (isBreakDecision) {
            FlowchartNode breakBranch    = trueEndsWithBreak ? node.getTrueBranch()  : node.getFalseBranch();
            FlowchartNode continueBranch = trueEndsWithBreak ? node.getFalseBranch() : node.getTrueBranch();
            String breakLabel    = trueEndsWithBreak ? "ДА"  : "НЕТ";
            String continueLabel = trueEndsWithBreak ? "НЕТ" : "ДА";

            // Use fixed BREAK_HORIZONTAL_SPACING — no clamping to returnRightX here,
            // because returnRightX will be computed from maxX AFTER body rendering.
            double colOffset = BREAK_HORIZONTAL_SPACING;

            double daColX  = x - colOffset;
            double netColX = x + colOffset;
            trackX(daColX  - PROCESS_WIDTH / 2);
            trackX(netColX + PROCESS_WIDTH / 2);

            double tipY      = y + halfH;
            double colStartY = y + h + VERTICAL_SPACING;

            labelText(breakLabel, x - halfW - 30, tipY - 10);
            line(x - halfW, tipY, daColX, tipY);
            arrow(daColX, tipY, daColX, colStartY - 5);
            double daColEndY = renderBreakColumnChain(breakBranch, daColX, colStartY);

            double netColEndY;
            if (continueBranch != null) {
                labelText(continueLabel, x + halfW + 10, tipY - 10);
                line(x + halfW, tipY, netColX, tipY);
                arrow(netColX, tipY, netColX, colStartY - 5);
                netColEndY = renderBreakColumnChain(continueBranch, netColX, colStartY);
            } else {
                labelText(continueLabel, x + halfW + 10, tipY - 10);
                line(x + halfW, tipY, netColX, tipY);
                arrow(netColX, tipY, netColX, colStartY - 5);
                netColEndY = colStartY;
            }

            breakGeometry.put(node, new double[]{daColX, netColX, daColEndY, netColEndY});

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

    private double renderBreakTailChain(List<FlowchartNode> tail, double x, double startY) {
        double currentY = startY;

        for (int i = 0; i < tail.size(); i++) {
            FlowchartNode node = tail.get(i);

            if (node instanceof ProcessNode) {
                rendered.add(node);
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

    private void drawDiamond(double x, double y, double w, double h) {
        double halfW = w / 2, halfH = h / 2;
        String points = String.format(Locale.US,
                "%.1f,%.1f %.1f,%.1f %.1f,%.1f %.1f,%.1f",
                x, y, x + halfW, y + halfH, x, y + h, x - halfW, y + halfH);
        svg.append(String.format(Locale.US, "<polygon class='shape' points='%s'/>\n", points));
    }

    private void line(double x1, double y1, double x2, double y2) {
        svg.append(String.format(Locale.US,
                "<line class=\"line\" x1=\"%.1f\" y1=\"%.1f\" x2=\"%.1f\" y2=\"%.1f\"/>\n",
                x1, y1, x2, y2));
    }

    private void arrow(double x1, double y1, double x2, double y2) {
        svg.append(String.format(Locale.US,
                "<line class=\"arrow\" x1=\"%.1f\" y1=\"%.1f\" x2=\"%.1f\" y2=\"%.1f\"/>\n",
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