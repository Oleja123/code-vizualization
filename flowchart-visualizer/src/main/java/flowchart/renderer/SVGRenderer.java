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
    private Map<DecisionNode, double[]> breakGeometry;    // daColX, netColX, daColEndY, netColEndY
    // continueGeometry: colX, colEndY  (the column where continue branch ends → arrow to loop top)
    private Map<DecisionNode, double[]> continueGeometry;
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
        continueGeometry = new HashMap<>();
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
            // Нет блоков после if — ищем терминал "конец" среди next-узлов веток
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

        // Запоминаем maxX до рендера тела — внешние колонки (if-ветки и т.д.) не должны влиять
        double maxXBeforeBody = maxX;

        // Render body chain
        List<FlowchartNode> bodyChain = collectBodyChain(node.getLoopBody(), node);
        double bodyEndY = renderLoopBodyChain(bodyChain, rightX, branchY, node);
        if (bodyEndY < 0) bodyEndY = -bodyEndY; // unwrap sentinel

        // --- COMPUTE returnRightX AFTER body render ---
        // Берём maxX накопленный телом цикла, но не меньше правой колонки тела.
        // Если тело не расширило maxX (внешние колонки были правее) — используем rightX + половина блока.
        double bodyOnlyMaxX = (maxX > maxXBeforeBody) ? maxX : rightX + PROCESS_WIDTH / 2;
        double returnRightX = bodyOnlyMaxX + BACK_ARROW_MARGIN;
        currentLoopReturnRightX = returnRightX;

        // НЕТ exits straight down from diamond bottom.
        double diamondBottom = y + h;
        labelText("НЕТ", x + 8, diamondBottom + 15);

        // maxCornerY: lowest Y of all back-arrow horizontal segments.
        // exitY for loop НЕТ branch must sit below this.
        double maxCornerY = bodyEndY + VERTICAL_SPACING / 2;

        // First pass: collect all corner-Y values so we can compute exitY before drawing
        for (double[] g : breakGeometry.values()) {
            double netColEndY = g[3];
            maxCornerY = Math.max(maxCornerY, netColEndY + VERTICAL_SPACING / 2);
            boolean isContinueDec = (g.length > 4 && g[4] == 1.0);
            if (isContinueDec) {
                double daColEndY = g[2];
                maxCornerY = Math.max(maxCornerY, daColEndY + VERTICAL_SPACING / 2);
            }
        }

        double exitY = maxCornerY + VERTICAL_SPACING / 2;

        // Second pass: draw all column connections
        for (double[] g : breakGeometry.values()) {
            double daColX     = g[0];
            double netColX    = g[1];
            double daColEndY  = g[2];
            double netColEndY = g[3];
            boolean isContinueDec = (g.length > 4 && g[4] == 1.0);

            if (isContinueDec) {
                // Continue-decision:
                //   НЕТ (right/normal col) → longer drop → right to returnRightX → up → arrow to diamond
                //   ДА  (left/continue col) → drop to MIDDLE of НЕТ line → arrow right to НЕТ column
                double netDropLength = VERTICAL_SPACING * 1.5; // longer drop for НЕТ
                double netCornerY    = netColEndY + netDropLength;
                double joinY         = netColEndY + netDropLength / 2; // middle of НЕТ line

                // НЕТ column: vertical down (longer) → right → up → arrow to mid of incoming arrow
                double contTargetY = y - VERTICAL_SPACING / 2;
                line(netColX, netColEndY, netColX, netCornerY);
                line(netColX, netCornerY, returnRightX, netCornerY);
                line(returnRightX, netCornerY, returnRightX, contTargetY);
                arrow(returnRightX, contTargetY, x + 5, contTargetY);
                trackX(netColX + 5);
                trackX(returnRightX + 5);

                // ДА column: drop to joinY (middle of НЕТ line), then arrow right to НЕТ column
                line(daColX, daColEndY, daColX, joinY);
                arrow(daColX, joinY, netColX - 1, joinY);
                trackX(daColX - 5);

            } else {
                // Break-decision:
                //   ДА (left col)  → down to breakJoinY → horizontal arrow to loop axis (exit)
                //   НЕТ (right col) → short drop → right to returnRightX → up → arrow to diamond
                double netCornerY = netColEndY + VERTICAL_SPACING / 2;
                double breakTargetY = y - VERTICAL_SPACING / 2;
                line(netColX, netColEndY, netColX, netCornerY);
                line(netColX, netCornerY, returnRightX, netCornerY);
                line(returnRightX, netCornerY, returnRightX, breakTargetY);
                arrow(returnRightX, breakTargetY, x + 5, breakTargetY);
                trackX(returnRightX + 5);

                double breakJoinY = exitY - VERTICAL_SPACING / 2;
                line(daColX, daColEndY, daColX, breakJoinY);
                arrow(daColX, breakJoinY, x - 1, breakJoinY);
                trackX(daColX - 5);
            }
        }

        // Old continueGeometry entries — now empty since continue uses breakGeometry,
        // but keep the loop for safety
        for (double[] g : continueGeometry.values()) { /* no-op */ }

        // Normal back-arrow: drawn when no break/continue-decisions (body returns via bodyEndY)
        if (breakGeometry.isEmpty()) {
            double backStartX  = Double.isNaN(lastBodyBlockX) ? rightX : lastBodyBlockX;
            double backCornerY = bodyEndY + VERTICAL_SPACING / 2;
            double backTargetY = y - VERTICAL_SPACING / 2; // mid of incoming arrow to diamond
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
        // Body start coordinates — needed for continue arrows to loop back here
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
                    // Last block in body — save its X so back-arrow starts from correct column
                    lastBodyBlockX = x;
                    return nodeBottom;
                }

            } else if (node instanceof DecisionNode decisionNode) {
                rendered.add(node);
                node.setPosition(x, currentY);

                double mergeY = renderDecisionInBody(decisionNode, x, currentY, nextNode, loop, bodyStartX, bodyStartY);

                if (mergeY < 0) {
                    double blockBottom = -mergeY;

                    if (nextNode != null) {
                        double[] g = breakGeometry.get(decisionNode);
                        if (g != null) {
                            // For both break and continue decisions: tail goes into NET (right) column
                            // break:    netColX = right = НЕТ (continues in body)
                            // continue: netColX = right = НЕТ (continues in body)
                            double tailColX   = g[1];
                            double tailStartY = g[3];
                            List<FlowchartNode> tail = chain.subList(i + 1, chain.size());
                            rendered.addAll(tail);
                            double tailEndY = renderBreakTailChain(tail, tailColX, tailStartY);
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
                                        double loopBodyX, double loopBodyY) {
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

        // ── CONTINUE DECISION ─────────────────────────────────────────────────
        // ДА (continue) → LEFT  column → connects to back-arrow (returnRightX)
        // НЕТ (normal)  → RIGHT column → body continues (tail nodes, then back-arrow)
        if (isContinueDecision && !isBreakDecision) {
            FlowchartNode contBranch   = trueEndsWithContinue ? node.getTrueBranch()  : node.getFalseBranch();
            FlowchartNode normalBranch = trueEndsWithContinue ? node.getFalseBranch() : node.getTrueBranch();
            String contLabel   = trueEndsWithContinue ? "ДА"  : "НЕТ";
            String normalLabel = trueEndsWithContinue ? "НЕТ" : "ДА";

            double colOffset = BREAK_HORIZONTAL_SPACING;
            double leftColX  = x - colOffset;  // ДА (continue) → LEFT → back-arrow
            double rightColX = x + colOffset;  // НЕТ (normal)  → RIGHT → body continues
            trackX(leftColX  - PROCESS_WIDTH / 2);
            trackX(rightColX + PROCESS_WIDTH / 2);

            double tipY      = y + halfH;
            double colStartY = y + h + VERTICAL_SPACING;

            // Continue (ДА) → LEFT column
            labelText(contLabel, x - halfW - 30, tipY - 10);
            line(x - halfW, tipY, leftColX, tipY);
            arrow(leftColX, tipY, leftColX, colStartY - 5);
            double leftColEndY = renderBreakColumnChain(contBranch, leftColX, colStartY);

            // Normal (НЕТ) → RIGHT column
            double rightColEndY;
            if (normalBranch != null) {
                labelText(normalLabel, x + halfW + 10, tipY - 10);
                line(x + halfW, tipY, rightColX, tipY);
                arrow(rightColX, tipY, rightColX, colStartY - 5);
                rightColEndY = renderBreakColumnChain(normalBranch, rightColX, colStartY);
            } else {
                labelText(normalLabel, x + halfW + 10, tipY - 10);
                line(x + halfW, tipY, rightColX, tipY);
                arrow(rightColX, tipY, rightColX, colStartY - 5);
                rightColEndY = colStartY;
            }

            // Store in breakGeometry with flag=1.0 (continue-decision):
            //   daColX  = leftColX  (ДА/continue) → back-arrow
            //   netColX = rightColX (НЕТ/normal)  → back-arrow (after tail nodes)
            breakGeometry.put(node, new double[]{leftColX, rightColX, leftColEndY, rightColEndY, 1.0});

            double blockBottom = Math.max(leftColEndY, rightColEndY) + VERTICAL_SPACING;
            node.setSize(w, blockBottom - y);
            updateMaxY(blockBottom);

            return -blockBottom;
        }

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
    //
    // Layout (matching the screenshot):
    //   1. Body nodes render top-down at column x (body start y = bodyStartY)
    //   2. Condition diamond renders BELOW the last body node
    //   3. "+" (true)  → left tip of diamond → left column → up back to bodyStartY
    //   4. "-" (false) → bottom of diamond → down → exitNode
    //
    private void renderDoWhile(DoWhileNode node, double x, double y, FlowchartNode stopBefore) {
        // arrowMidY = y: внешняя стрелка приходит сюда с наконечником,
        // стрелка возврата (ДА) тоже сюда — один общий наконечник.
        // Входящая стрелка снаружи идёт от (y - VERTICAL_SPACING) до y.
        // Стрелка возврата ДА присоединяется к середине этой стрелки: y - VERTICAL_SPACING/2.
        double arrowMidY = y - VERTICAL_SPACING / 2;
        double firstBlockY = y;
        double bodyStartY = firstBlockY;

        // ── Render body chain top-down (используем renderLoopBodyChain — он умеет DecisionNode) ──
        // Создаём временный LoopStartNode-совместимый обход через collectBodyChain/renderLoopBodyChain.
        // Для do-while нет LoopStartNode, поэтому собираем цепочку вручную через collectDoWhileBodyChain,
        // но рендерим каждый узел через renderNode чтобы поддержать DecisionNode внутри тела.
        double bodyEndY = renderDoWhileBody(node.getLoopBody(), node, x, firstBlockY);
        if (bodyEndY < 0) bodyEndY = -bodyEndY;

        // ── Arrow from last body node to diamond ──
        double dY = bodyEndY + VERTICAL_SPACING;
        arrow(x, bodyEndY, x, dY - 5);

        // ── Condition diamond below body ──
        double dW = DECISION_WIDTH, dH = DECISION_HEIGHT;
        double halfW = dW / 2, halfH = dH / 2;
        node.setPosition(x, dY);
        drawDiamond(x, dY, dW, dH);
        text(node.getLabel(), x, dY + halfH);
        updateMaxY(dY + dH);

        // ── "ДА" back-arrow: left tip → left column → середина стрелки к первому блоку тела ──
        // Используем minX тела - отступ, чтобы не наслаиваться на ветки if внутри тела
        double leftColX = minX - BACK_ARROW_MARGIN;
        double tipY = dY + halfH;
        double leftTipX = x - halfW;

        line(leftTipX, tipY, leftColX, tipY);
        line(leftColX, tipY, leftColX, arrowMidY);
        arrow(leftColX, arrowMidY, x - 1, arrowMidY);
        labelText("ДА", leftTipX - 40, tipY - 10);
        trackX(leftColX - 5);

        // ── "НЕТ" (false) exit arrow: diamond bottom → down ──
        double dBottom = dY + dH;
        labelText("НЕТ", x + 8, dBottom + 15);

        if (node.getExitNode() != null) {
            FlowchartNode exitNode = node.getExitNode();
            if (exitNode instanceof TerminalNode t && !t.isStart()) {
                // exitNode — это сам терминал "конец"
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
            // Нет exitNode — конец программы (не должно происходить после fix в builder)
            endArrowFromX = x;
            endArrowFromY = dBottom;
            node.setSize(dW, dBottom - y);
        }
    }

    /**
     * Рендерит тело do-while сверху вниз, поддерживая ProcessNode и DecisionNode.
     * Возвращает Y нижней границы последнего нарисованного узла.
     */
    private double renderDoWhileBody(FlowchartNode start, DoWhileNode loop, double x, double startY) {
        double currentY = startY;
        FlowchartNode cur = start;
        IdentityHashMap<FlowchartNode, Boolean> seen = new IdentityHashMap<>();

        while (cur != null && !seen.containsKey(cur)) {
            if (cur instanceof LoopEndNode)   break;
            if (cur instanceof TerminalNode)  break;
            if (cur == loop.getExitNode())    break;

            seen.put(cur, true);
            rendered.add(cur);
            cur.setPosition(x, currentY);

            if (cur instanceof ProcessNode) {
                renderProcess(cur, x, currentY);
                updateMaxY(currentY + PROCESS_HEIGHT);
                trackX(x + PROCESS_WIDTH / 2);
                double nodeBottom = currentY + PROCESS_HEIGHT;

                // Найти следующий узел тела
                FlowchartNode next = null;
                for (FlowchartNode n : cur.getNext()) {
                    if (n instanceof LoopEndNode)     continue;
                    if (n instanceof TerminalNode)    continue;
                    if (n == loop.getExitNode())      continue;
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

            } else if (cur instanceof DecisionNode decNode) {
                // Рендерим if внутри тела do-while как обычный DecisionNode
                // Найти следующий узел после if (после слияния веток)
                FlowchartNode afterIf = null;
                for (FlowchartNode n : decNode.getNext()) {
                    if (n instanceof LoopEndNode)     continue;
                    if (n instanceof TerminalNode)    continue;
                    if (n == loop.getExitNode())      continue;
                    if (n == decNode.getTrueBranch()) continue;
                    if (n == decNode.getFalseBranch()) continue;
                    afterIf = n;
                    break;
                }

                renderDecision(decNode, x, currentY, afterIf);
                double decBottom = currentY + decNode.getHeight();

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

    private List<FlowchartNode> collectDoWhileBodyChain(FlowchartNode start, DoWhileNode loop) {
        List<FlowchartNode> chain = new ArrayList<>();
        IdentityHashMap<FlowchartNode, Boolean> seen = new IdentityHashMap<>();
        FlowchartNode cur = start;
        while (cur != null && !seen.containsKey(cur)) {
            if (cur instanceof LoopEndNode)   break;
            if (cur instanceof TerminalNode)  break;
            if (cur == loop.getExitNode())    break;
            // Only collect ProcessNodes for now (simple do-while body)
            if (!(cur instanceof ProcessNode)) break;
            seen.put(cur, true);
            chain.add(cur);
            FlowchartNode next = null;
            for (FlowchartNode n : cur.getNext()) {
                if (n instanceof LoopEndNode)      continue;
                if (n instanceof TerminalNode)     continue;
                if (n == loop.getExitNode())       continue;
                next = n;
                break;
            }
            cur = next;
        }
        return chain;
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