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
    private boolean endArrowAlreadyDrawn;

    private double currentLoopReturnRightX = Double.MAX_VALUE;
    private double lastBodyBlockX = Double.NaN;
    private double lastBodyBlockBottomY = Double.NaN;

    private List<Object[]> deferredDoWhileContinueCols;
    private List<double[]> doWhileBreakExits;
    private List<double[]> loopReturnExits;

    // ─────────────────────────────────────────────────────────────────────────
    // Вспомогательные методы для обёртки узлов в <g data-line="...">
    // ─────────────────────────────────────────────────────────────────────────

    private void beginNodeGroup(FlowchartNode node) {
        Location loc = node.getAstLocation();
        if (loc != null) {
            svg.append(String.format(
                    "<g data-node-id=\"%s\" data-line=\"%d\" data-line-end=\"%d\">%n",
                    escapeXml(node.getId()),
                    loc.getLine(),
                    loc.getEndLine()
            ));
        } else {
            svg.append(String.format("<g data-node-id=\"%s\">%n", escapeXml(node.getId())));
        }
    }

    private void endNodeGroup() {
        svg.append("</g>\n");
    }

    // ─────────────────────────────────────────────────────────────────────────

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
        endArrowAlreadyDrawn = false;
        currentLoopReturnRightX = Double.MAX_VALUE;
        lastBodyBlockX = Double.NaN;
        lastBodyBlockBottomY = Double.NaN;
        deferredDoWhileContinueCols = new ArrayList<>();
        doWhileBreakExits = new ArrayList<>();
        loopReturnExits = new ArrayList<>();

        renderNode(start, 700, 100, null);

        double padding = 60;

        if (endNode != null) {
            double endX = endArrowFromX;
            double endY;
            if (endArrowAlreadyDrawn) {
                endY = endArrowFromY;
            } else {
                endY = endArrowFromY + VERTICAL_SPACING;
                arrow(endX, endArrowFromY, endX, endY);
            }
            renderTerminal(endNode, endX, endY);
            updateMaxY(endY + TERMINAL_HEIGHT);

            if (!loopReturnExits.isEmpty()) {
                double mergeY = endArrowFromY + VERTICAL_SPACING / 2;
                for (double[] exit : loopReturnExits) {
                    double exitX      = exit[0];
                    double exitBottomY = exit[1];
                    line(exitX, exitBottomY, exitX, mergeY);
                    arrow(exitX, mergeY, endX + 1, mergeY);
                }
            }
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
        out.append(".shape  { fill: white; stroke: black; stroke-width: 2; transition: fill 0.25s, stroke 0.25s; }\n");
        out.append(".line   { stroke: black; stroke-width: 2; fill: none; }\n");
        out.append(".arrow  { stroke: black; stroke-width: 2; fill: none; marker-end: url(#arrow); }\n");
        out.append(".text   { font-family: Arial; font-size: 13px; text-anchor: middle; dominant-baseline: middle; }\n");
        out.append(".label  { font-family: Arial; font-size: 11px; fill: #333; }\n");
        // Стили подсветки активного блока при трассировке
        out.append(".node-active > .shape  { fill: #fff9c4 !important; stroke: #f59e0b !important; stroke-width: 3 !important; }\n");
        out.append(".node-active > text    { font-weight: bold; }\n");
        out.append("</style>\n");
        out.append("</defs>\n");

        out.append(svg);
        out.append("</svg>");
        return out.toString();
    }

    public Map<String, String> renderAll(Map<String, FlowchartNode> functions) {
        Map<String, String> result = new java.util.LinkedHashMap<>();
        List<String> names = sortedFunctionNames(functions.keySet());
        for (String name : names) {
            result.put(name, render(functions.get(name)));
        }
        return result;
    }

    public String renderAllInOne(Map<String, FlowchartNode> functions) {
        List<String> names = sortedFunctionNames(functions.keySet());
        if (names.isEmpty()) return "<svg xmlns=\"http://www.w3.org/2000/svg\"/>";
        if (names.size() == 1) return render(functions.get(names.get(0)));

        double GAP_BETWEEN   = 80;
        double LABEL_HEIGHT  = 30;
        double padding       = 60;

        List<Object[]> parts = new ArrayList<>();

        for (String name : names) {
            String svgStr = render(functions.get(name));
            double[] vb = parseViewBox(svgStr);
            String inner = extractInnerContent(svgStr);
            parts.add(new Object[]{name, inner, vb[0], vb[1], vb[2], vb[3]});
        }

        double totalW = 0;
        double maxH   = 0;
        for (Object[] p : parts) {
            double vbW = (double) p[4];
            double vbH = (double) p[5];
            totalW += vbW;
            if (vbH + LABEL_HEIGHT > maxH) maxH = vbH + LABEL_HEIGHT;
        }
        totalW += GAP_BETWEEN * (parts.size() - 1);

        double svgW = totalW;
        double svgH = maxH + padding;

        StringBuilder out = new StringBuilder();
        out.append("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n");
        out.append(String.format(Locale.US,
                "<svg xmlns=\"http://www.w3.org/2000/svg\" " +
                        "width=\"100%%\" height=\"100%%\" " +
                        "viewBox=\"0 0 %.1f %.1f\" " +
                        "preserveAspectRatio=\"xMidYMin meet\">\n",
                svgW, svgH));

        out.append("<defs>\n");
        out.append("<marker id=\"arrow\" markerWidth=\"10\" markerHeight=\"10\" refX=\"9\" refY=\"5\" orient=\"auto\">\n");
        out.append("<path d=\"M0,0 L10,5 L0,10 z\" fill=\"black\"/>\n");
        out.append("</marker>\n");
        out.append("<style>\n");
        out.append(".shape  { fill: white; stroke: black; stroke-width: 2; transition: fill 0.25s, stroke 0.25s; }\n");
        out.append(".line   { stroke: black; stroke-width: 2; fill: none; }\n");
        out.append(".arrow  { stroke: black; stroke-width: 2; fill: none; marker-end: url(#arrow); }\n");
        out.append(".text   { font-family: Arial; font-size: 13px; text-anchor: middle; dominant-baseline: middle; }\n");
        out.append(".label  { font-family: Arial; font-size: 11px; fill: #333; }\n");
        out.append(".func-label { font-family: Arial; font-size: 15px; font-weight: bold; fill: #444; text-anchor: middle; }\n");
        out.append(".node-active > .shape  { fill: #fff9c4 !important; stroke: #f59e0b !important; stroke-width: 3 !important; }\n");
        out.append(".node-active > text    { font-weight: bold; }\n");
        out.append("</style>\n");
        out.append("</defs>\n");

        double curX = 0;
        for (int i = 0; i < parts.size(); i++) {
            Object[] p    = parts.get(i);
            String pName  = (String) p[0];
            String pInner = (String) p[1];
            double pVbX   = (double) p[2];
            double pVbY   = (double) p[3];
            double pVbW   = (double) p[4];
            double pVbH   = (double) p[5];

            double labelX = curX + pVbW / 2;
            double labelY = LABEL_HEIGHT / 2 + 5;
            out.append(String.format(Locale.US,
                    "<text class=\"func-label\" x=\"%.1f\" y=\"%.1f\">%s</text>\n",
                    labelX, labelY, escapeXml(pName)));

            double tx = curX - pVbX;
            double ty = LABEL_HEIGHT - pVbY;
            out.append(String.format(Locale.US,
                    "<g transform=\"translate(%.3f, %.3f)\">\n", tx, ty));
            out.append(pInner);
            out.append("</g>\n");

            if (i < parts.size() - 1) {
                double sepX = curX + pVbW + GAP_BETWEEN / 2;
                out.append(String.format(Locale.US,
                        "<line x1=\"%.1f\" y1=\"%.1f\" x2=\"%.1f\" y2=\"%.1f\" " +
                                "style=\"stroke:#ccc;stroke-width:1;stroke-dasharray:6,4\"/>\n",
                        sepX, 0.0, sepX, svgH));
            }

            curX += pVbW + GAP_BETWEEN;
        }

        out.append("</svg>");
        return out.toString();
    }

    private double[] parseViewBox(String svg) {
        int idx = svg.indexOf("viewBox=\"");
        if (idx < 0) return new double[]{0, 0, 800, 600};
        int start = idx + 9;
        int end   = svg.indexOf('"', start);
        String[] parts = svg.substring(start, end).trim().split("\\s+");
        return new double[]{
                Double.parseDouble(parts[0]),
                Double.parseDouble(parts[1]),
                Double.parseDouble(parts[2]),
                Double.parseDouble(parts[3])
        };
    }

    private String extractInnerContent(String svgStr) {
        int defsEnd = svgStr.indexOf("</defs>");
        if (defsEnd < 0) {
            int svgOpen = svgStr.indexOf('>') + 1;
            int svgClose = svgStr.lastIndexOf("</svg>");
            return svgStr.substring(svgOpen, svgClose);
        }
        int contentStart = defsEnd + "</defs>".length();
        int svgClose = svgStr.lastIndexOf("</svg>");
        return svgStr.substring(contentStart, svgClose);
    }

    private List<String> sortedFunctionNames(java.util.Collection<String> names) {
        List<String> sorted = new ArrayList<>(names);
        sorted.sort((a, b) -> {
            if (a.equals("main")) return -1;
            if (b.equals("main")) return 1;
            return a.compareTo(b);
        });
        return sorted;
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
        beginNodeGroup(node);
        svg.append(String.format(Locale.US,
                "<ellipse class=\"shape\" cx=\"%.1f\" cy=\"%.1f\" rx=\"%.1f\" ry=\"%.1f\"/>\n",
                x, y + h / 2, w / 2, h / 2));
        textWrapped(node.getLabel(), x, y + h / 2, w * 0.85, 16);
        endNodeGroup();
        trackX(x - w / 2);
        trackX(x + w / 2);
    }

    private void renderProcess(FlowchartNode node, double x, double y) {
        double maxW = PROCESS_WIDTH - 16;
        List<String> lines = wrapText(node.getLabel(), maxW, 13);
        double lineH   = 18;
        double minH    = PROCESS_HEIGHT;
        double actualH = Math.max(minH, lines.size() * lineH + 20);

        node.setSize(PROCESS_WIDTH, actualH);
        beginNodeGroup(node);
        svg.append(String.format(Locale.US,
                "<rect class=\"shape\" x=\"%.1f\" y=\"%.1f\" width=\"%.1f\" height=\"%.1f\"/>\n",
                x - PROCESS_WIDTH / 2, y, PROCESS_WIDTH, actualH));

        double textStartY = y + actualH / 2 - (lines.size() - 1) * lineH / 2.0;
        for (int i = 0; i < lines.size(); i++) {
            double lineY = textStartY + i * lineH;
            svg.append(String.format(Locale.US,
                    "<text class=\"text\" x=\"%.1f\" y=\"%.1f\">%s</text>\n",
                    x, lineY, escapeXml(lines.get(i))));
        }
        endNodeGroup();
    }

    private void renderDecision(DecisionNode node, double x, double y, FlowchartNode stopBefore) {
        double w = DECISION_WIDTH;
        double h = diamondHeight(node.getLabel());
        double halfW = w / 2, halfH = h / 2;

        node.setSize(w, h);
        beginNodeGroup(node);
        drawDiamondWithText(x, y, w, h, node.getLabel());
        endNodeGroup();

        boolean trueIsReturn  = chainEndsWithReturn(node.getTrueBranch());
        boolean falseIsReturn = chainEndsWithReturn(node.getFalseBranch());
        boolean hasReturn     = trueIsReturn || falseIsReturn;
        boolean onlyOneReturn = hasReturn && !(trueIsReturn && falseIsReturn);

        if (onlyOneReturn) {
            FlowchartNode returnBranch = trueIsReturn ? node.getTrueBranch() : node.getFalseBranch();
            FlowchartNode elseBranch   = trueIsReturn ? node.getFalseBranch() : node.getTrueBranch();
            String returnLabel = trueIsReturn ? "ДА" : "НЕТ";
            String elseLabel   = trueIsReturn ? "НЕТ" : "ДА";

            double rightTipX  = x + halfW;
            double rightTipY  = y + halfH;
            double bottomTipY = y + h;

            double retColX   = x + HORIZONTAL_SPACING;
            double colStartY = bottomTipY + VERTICAL_SPACING;
            trackX(retColX + PROCESS_WIDTH / 2);
            labelText(returnLabel, rightTipX + 10, rightTipY - 10);
            line(rightTipX, rightTipY, retColX, rightTipY);

            double retColEndY;
            if (isBreak(returnBranch)) {
                retColEndY = rightTipY;
            } else {
                arrow(retColX, rightTipY, retColX, colStartY);
                retColEndY = renderBreakColumnChain(returnBranch, retColX, colStartY, null);
            }

            double exitFromY = (retColEndY <= rightTipY + 1) ? rightTipY : retColEndY;
            loopReturnExits.add(new double[]{retColX, exitFromY});

            labelText(elseLabel, x + 8, bottomTipY + 15);

            double blockH = Math.max(h, retColEndY - y + VERTICAL_SPACING);
            node.setSize(w, blockH);
            updateMaxY(y + blockH);

            FlowchartNode afterIf = null;
            for (FlowchartNode n : node.getNext()) {
                if (n != returnBranch && n != elseBranch) { afterIf = n; break; }
            }

            if (elseBranch != null && !isBreak(elseBranch)) {
                double elseStartY = bottomTipY + VERTICAL_SPACING;
                arrow(x, bottomTipY, x, elseStartY);
                double elseEndY = renderBreakColumnChain(elseBranch, x, elseStartY, null, afterIf);
                if (afterIf != null && !rendered.contains(afterIf)) {
                    double nextY = elseEndY + VERTICAL_SPACING;
                    arrow(x, elseEndY, x, nextY - 5);
                    renderNode(afterIf, x, nextY, stopBefore);
                } else {
                    endArrowFromX = x;
                    endArrowFromY = elseEndY;
                }
            } else {
                if (afterIf != null && !rendered.contains(afterIf)) {
                    double nextY = bottomTipY + VERTICAL_SPACING;
                    arrow(x, bottomTipY, x, nextY - 5);
                    renderNode(afterIf, x, nextY, stopBefore);
                } else {
                    endArrowFromX = x;
                    endArrowFromY = bottomTipY;
                }
            }
            return;
        }

        double branchY = y + h + VERTICAL_SPACING;
        double leftX   = x - HORIZONTAL_SPACING;
        trackX(leftX - PROCESS_WIDTH / 2);

        double leftBottom  = branchY;
        double rightBottom = branchY;

        if (node.getTrueBranch() != null) {
            line(x - halfW, y + halfH, leftX, y + halfH);
            line(leftX, y + halfH, leftX, branchY - 5);
            arrow(leftX, branchY - 5, leftX, branchY);
            labelText("ДА", x - halfW - 30, y + halfH - 10);

            renderNode(node.getTrueBranch(), leftX, branchY, stopBefore);
            leftBottom = branchY + branchHeight(node.getTrueBranch(), stopBefore);

            double minRightX = x + HORIZONTAL_SPACING;
            double rightX = Math.max(minRightX, maxX + HORIZONTAL_SPACING / 2);

            if (node.getFalseBranch() != null) {
                trackX(rightX + PROCESS_WIDTH / 2);
                line(x + halfW, y + halfH, rightX, y + halfH);
                arrow(rightX, y + halfH, rightX, branchY - 5);
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
                trackX(noElseRightX + 5);
                line(x + halfW, y + halfH, noElseRightX, y + halfH);
                labelText("НЕТ", x + halfW + 10, y + halfH - 10);
                line(noElseRightX, y + halfH, noElseRightX, mergeY);
                line(noElseRightX, mergeY, x, mergeY);
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
            return;
        }

        double rightX = x + HORIZONTAL_SPACING;
        trackX(rightX + PROCESS_WIDTH / 2);

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

        if (node.getFalseBranch() != null) {
            line(rightX, rightBottom, rightX, mergeY);
            line(rightX, mergeY, x, mergeY);
        } else {
            line(x + halfW, y + halfH, rightX, y + halfH);
            labelText("НЕТ", x + halfW + 10, y + halfH - 10);
            line(rightX, y + halfH, rightX, mergeY);
            line(rightX, mergeY, x, mergeY);
            trackX(rightX + 5);
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

    private void renderLoop(LoopStartNode node, double x, double y, FlowchartNode stopBefore) {
        double w = DECISION_WIDTH;
        double h = diamondHeight(node.getLabel());
        double halfW = w / 2, halfH = h / 2;

        double incomingArrowMidY = y - VERTICAL_SPACING / 2;

        beginNodeGroup(node);
        drawDiamondWithText(x, y, w, h, node.getLabel());
        endNodeGroup();

        double rightX  = x + HORIZONTAL_SPACING + PROCESS_WIDTH / 2 + BACK_ARROW_MARGIN / 2;
        double branchY = y + h + VERTICAL_SPACING;

        line(x + halfW, y + halfH, rightX, y + halfH);
        arrow(rightX, y + halfH, rightX, branchY - 5);
        labelText("ДА", x + halfW + 10, y + halfH - 10);

        currentLoopReturnRightX = Double.MAX_VALUE;
        lastBodyBlockX = Double.NaN;
        lastBodyBlockBottomY = Double.NaN;

        double savedEndArrowFromX = endArrowFromX;
        double savedEndArrowFromY = endArrowFromY;

        List<FlowchartNode> bodyChain = collectBodyChain(node.getLoopBody(), node);
        double bodyEndY = renderLoopBodyChain(bodyChain, rightX, branchY, node);
        if (bodyEndY < 0) bodyEndY = -bodyEndY;

        endArrowFromX = savedEndArrowFromX;
        endArrowFromY = savedEndArrowFromY;

        double returnRightX = maxX + BACK_ARROW_MARGIN;
        currentLoopReturnRightX = returnRightX;

        double diamondBottom = y + h;
        labelText("НЕТ", x + 8, diamondBottom + 15);

        double maxCornerY = bodyEndY + VERTICAL_SPACING / 2;
        for (double[] g : breakGeometry.values()) {
            boolean isContinueDec = (g.length > 4 && g[4] == 1.0);
            if (isContinueDec) continue;
            boolean isNewFmt = (g.length > 10 && g[10] == 1.0);
            if (isNewFmt) {
                double breakEndY = g[2];
                double tipY      = (g.length > 5) ? g[5] : 0;
                boolean isBare   = (breakEndY <= tipY + 1);
                double effectiveEndY = isBare ? tipY : breakEndY;
                maxCornerY = Math.max(maxCornerY, effectiveEndY + VERTICAL_SPACING / 2);
            } else {
                double netColEndY = g[3];
                maxCornerY = Math.max(maxCornerY, netColEndY + VERTICAL_SPACING / 2);
            }
        }

        double exitY = maxCornerY + VERTICAL_SPACING / 2;

        double contTargetY = incomingArrowMidY;

        for (double[] g : breakGeometry.values()) {
            double daColX     = g[0];
            double netColX    = g[1];
            double daColEndY  = g[2];
            double netColEndY = g[3];
            boolean isContinueDec  = (g.length > 4  && g[4] == 1.0);
            double tipY            = (g.length > 5)  ? g[5] : 0;
            double colStartY       = (g.length > 6)  ? g[6] : tipY + DECISION_HEIGHT / 2 + VERTICAL_SPACING;
            double routingNetColEndY = (g.length > 8) ? g[8] : netColEndY;
            boolean isReturnBreak  = (g.length > 9  && g[9] == 1.0);
            boolean isNewFormat    = (g.length > 10 && g[10] == 1.0);

            boolean daBare  = (daColEndY  <= colStartY + 1);
            boolean netBare = (routingNetColEndY <= colStartY + 1);

            if (isNewFormat) {
                double breakColX   = daColX;
                double breakEndY   = daColEndY;
                boolean isBare     = (breakEndY <= tipY + 1);

                if (isReturnBreak) {
                    double exitFromY = isBare ? tipY : breakEndY;
                    loopReturnExits.add(new double[]{breakColX, exitFromY});
                    trackX(breakColX + 5);
                } else {
                    double breakJoinY = exitY - VERTICAL_SPACING / 2;
                    double exitFromY  = isBare ? tipY : breakEndY;
                    line(breakColX, exitFromY, breakColX, breakJoinY);
                    arrow(breakColX, breakJoinY, x + 1, breakJoinY);
                    trackX(breakColX + 5);
                    if (breakJoinY + VERTICAL_SPACING / 2 > exitY) exitY = breakJoinY + VERTICAL_SPACING / 2;
                }
                continue;
            }

            if (isContinueDec) {
                double tailBottom    = netColEndY;
                boolean hasTailInNet = tailBottom > colStartY + 1;
                boolean netHasBlocks = !netBare || hasTailInNet;
                boolean daHasBlocks  = !daBare;

                double netBlockRightX, netMidRightY;
                if (!netBare) {
                    netBlockRightX = netColX + PROCESS_WIDTH / 2;
                    netMidRightY   = routingNetColEndY - PROCESS_HEIGHT / 2;
                } else if (hasTailInNet) {
                    netBlockRightX = netColX + PROCESS_WIDTH / 2;
                    netMidRightY   = tailBottom - PROCESS_HEIGHT / 2;
                } else {
                    netBlockRightX = netColX;
                    netMidRightY   = tipY;
                }

                if (netHasBlocks) {
                    line(netBlockRightX, netMidRightY, returnRightX, netMidRightY);
                    line(returnRightX, netMidRightY, returnRightX, contTargetY);
                    arrow(returnRightX, contTargetY, x + 1, contTargetY);
                    trackX(returnRightX + 5);
                } else {
                    line(netColX, tipY, returnRightX, tipY);
                    line(returnRightX, tipY, returnRightX, contTargetY);
                    arrow(returnRightX, contTargetY, x + 1, contTargetY);
                    trackX(netColX + 5);
                    trackX(returnRightX + 5);
                }

                double joinX = netHasBlocks
                        ? (netBlockRightX + returnRightX) / 2
                        : (netColX + returnRightX) / 2;
                double joinY = netHasBlocks ? netMidRightY : tipY;

                if (daHasBlocks) {
                    double loopDownY = Math.max(daColEndY, joinY) + VERTICAL_SPACING;
                    line(daColX, daColEndY, daColX, loopDownY);
                    line(daColX, loopDownY, joinX, loopDownY);
                    arrow(joinX, loopDownY, joinX, joinY + 1);
                    trackX(daColX - 5);
                    updateMaxY(loopDownY);
                } else {
                    double daLeftX   = daColX - BACK_ARROW_MARGIN;
                    double loopDownY = Math.max(tipY + DECISION_HEIGHT / 2 + VERTICAL_SPACING,
                            joinY + VERTICAL_SPACING / 2);
                    line(daColX, tipY, daLeftX, tipY);
                    line(daLeftX, tipY, daLeftX, loopDownY);
                    line(daLeftX, loopDownY, joinX, loopDownY);
                    arrow(joinX, loopDownY, joinX, joinY + 1);
                    trackX(daLeftX - 5);
                    updateMaxY(loopDownY);
                }

                {
                    double tailBottom2 = netColEndY;
                    boolean hasTail2 = tailBottom2 > colStartY + 1;
                    double contLoopDownY = daHasBlocks
                            ? Math.max(daColEndY, joinY) + VERTICAL_SPACING
                            : Math.max(
                            tipY + DECISION_HEIGHT / 2 + VERTICAL_SPACING,
                            Math.max(daBare ? 0 : daColEndY + VERTICAL_SPACING / 2,
                                    netBare
                                            ? (hasTail2
                                            ? Math.max(daColEndY, tailBottom2) + VERTICAL_SPACING / 2
                                            : 0)
                                            : routingNetColEndY + VERTICAL_SPACING / 2)
                    );
                    if (contLoopDownY + VERTICAL_SPACING / 2 > exitY) exitY = contLoopDownY + VERTICAL_SPACING / 2;
                }

            } else if (isReturnBreak) {
                double blockBottomY = daBare ? tipY : daColEndY;
                loopReturnExits.add(new double[]{daColX, blockBottomY});
                trackX(daColX - 5);

            } else {
                double tailEndY    = netColEndY;
                boolean hasTailInNet = tailEndY > colStartY + 1;
                double backTargetY = incomingArrowMidY;

                double netCornerY;
                if (hasTailInNet) {
                    double netMidRightX = netColX + PROCESS_WIDTH / 2;
                    double netMidRightY = tailEndY - PROCESS_HEIGHT / 2;
                    line(netMidRightX, netMidRightY, returnRightX, netMidRightY);
                    line(returnRightX, netMidRightY, returnRightX, backTargetY);
                    arrow(returnRightX, backTargetY, x + 1, backTargetY);
                    trackX(returnRightX + 5);
                    netCornerY = netMidRightY;
                    if (netCornerY + VERTICAL_SPACING / 2 > exitY) {
                        exitY = netCornerY + VERTICAL_SPACING / 2;
                    }
                } else {
                    netCornerY = tipY;
                    line(netColX, tipY, returnRightX, tipY);
                    line(returnRightX, netCornerY, returnRightX, backTargetY);
                    arrow(returnRightX, backTargetY, x + 1, backTargetY);
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

        if (!Double.isNaN(lastBodyBlockX)) {
            boolean hasNonReturnBreaks = false;
            for (double[] g : breakGeometry.values()) {
                boolean isContinueDec = (g.length > 4 && g[4] == 1.0);
                boolean isReturnBreak = (g.length > 9 && g[9] == 1.0);
                if (!isContinueDec && !isReturnBreak) {
                    hasNonReturnBreaks = true;
                    break;
                }
            }

            if (!hasNonReturnBreaks) {
                double backStartX  = lastBodyBlockX;
                double blockBottom = Double.isNaN(lastBodyBlockBottomY) ? bodyEndY : lastBodyBlockBottomY;
                double backTargetY = incomingArrowMidY;
                double rightEdgeX  = backStartX + PROCESS_WIDTH / 2;
                double midRightY   = blockBottom - PROCESS_HEIGHT / 2;
                line(rightEdgeX, midRightY, returnRightX, midRightY);
                line(returnRightX, midRightY, returnRightX, backTargetY);
                arrow(returnRightX, backTargetY, x + 1, backTargetY);
                trackX(returnRightX + 5);
            }
        }

        updateMaxY(bodyEndY);

        if (node.getExitNode() != null) {
            if (node.getExitNode() instanceof TerminalNode t && !t.isStart()) {
                line(x, diamondBottom, x, exitY);
                endNode = t;
                endArrowFromX = x;
                endArrowFromY = exitY;
                endArrowAlreadyDrawn = false;
                node.setSize(w, exitY - y);
                return;
            }

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

                if (exitNext instanceof ConnectorNode conn && "return".equals(conn.getLabel())) {
                    rendered.add(conn);
                    for (FlowchartNode returnNext : conn.getNext()) {
                        exitNext = returnNext;
                        break;
                    }
                }

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
            endArrowFromX = x;
            endArrowFromY = exitY;
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
                lastBodyBlockX = x;
                lastBodyBlockBottomY = nodeBottom;

                if (nextNode != null) {
                    double nextY = nodeBottom + VERTICAL_SPACING;
                    arrow(x, nodeBottom, x, nextY - 5);
                    currentY = nextY;
                } else {
                    return nodeBottom;
                }

            } else if (node instanceof DecisionNode decisionNode) {
                rendered.add(node);
                node.setPosition(x, currentY);

                List<FlowchartNode> tail = (i + 1 < chain.size()) ? chain.subList(i + 1, chain.size()) : List.of();
                rendered.addAll(tail);

                double mergeY = renderDecisionInBody(decisionNode, x, currentY, nextNode, loop, bodyStartX, bodyStartY, null, null);

                if (mergeY < 0) {
                    double blockBottom = -mergeY;

                    if (nextNode != null) {
                        double[] g = breakGeometry.get(decisionNode);
                        if (g != null) {
                            boolean isNewFmt = (g.length > 10 && g[10] == 1.0);

                            if (isNewFmt) {
                                double bottomX    = g[1];
                                double bottomTipY = g[3];
                                double nextY = bottomTipY + VERTICAL_SPACING;
                                arrow(bottomX, bottomTipY, bottomX, nextY - 5);
                                currentY = nextY;
                                blockBottom = bottomTipY;
                            } else {
                                double tailColX   = g[1];
                                double colStartY  = g[6];
                                double origNetEnd = g[3];
                                boolean columnWasBare = (origNetEnd <= colStartY + 1);

                                double[] extended = Arrays.copyOf(g, Math.max(11, g.length));
                                extended[8] = origNetEnd;
                                breakGeometry.put(decisionNode, extended);
                                g = extended;

                                if (columnWasBare) {
                                    arrow(tailColX, g[5], tailColX, colStartY);
                                    double tailEndY = renderTailBlocks(tail, tailColX, colStartY, false);
                                    g[3] = tailEndY;
                                    blockBottom = Math.max(-mergeY, tailEndY);
                                } else {
                                    double tailEndY = renderTailBlocks(tail, tailColX, origNetEnd, true);
                                    g[3] = tailEndY;
                                    blockBottom = Math.max(-mergeY, tailEndY);
                                }
                            }
                        }
                    }
                    if (!((breakGeometry.get(decisionNode) != null && breakGeometry.get(decisionNode).length > 10 && breakGeometry.get(decisionNode)[10] == 1.0) && nextNode != null)) {
                        return -blockBottom;
                    }
                    continue;
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
        double w = DECISION_WIDTH;
        double h = diamondHeight(node.getLabel());
        double halfW = w / 2, halfH = h / 2;
        node.setSize(w, h);
        beginNodeGroup(node);
        drawDiamondWithText(x, y, w, h, node.getLabel());
        endNodeGroup();

        boolean trueEndsWithBreak    = chainEndsWithBreak(node.getTrueBranch());
        boolean falseEndsWithBreak   = chainEndsWithBreak(node.getFalseBranch());
        boolean trueEndsWithContinue = chainEndsWithContinue(node.getTrueBranch());
        boolean falseEndsWithContinue= chainEndsWithContinue(node.getFalseBranch());
        boolean isBreakDecision      = trueEndsWithBreak  || falseEndsWithBreak;
        boolean isContinueDecision   = trueEndsWithContinue || falseEndsWithContinue;

        boolean insideDoWhile = (doWhileLoop != null);

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
                double colOffset = BREAK_HORIZONTAL_SPACING;
                double leftColX  = x - colOffset;
                double rightColX = x + colOffset;
                trackX(leftColX  - PROCESS_WIDTH / 2);
                trackX(rightColX + PROCESS_WIDTH / 2);

                double tipY      = y + halfH;
                double colStartY = y + h + VERTICAL_SPACING;

                labelText(contLabel, x - halfW - 30, tipY - 10);
                line(x - halfW, tipY, leftColX, tipY);

                double leftColEndY;
                if (isContinue(contBranch)) {
                    leftColEndY = colStartY;
                } else {
                    arrow(leftColX, tipY, leftColX, colStartY);
                    leftColEndY = renderBreakColumnChain(contBranch, leftColX, colStartY, loop);
                }

                labelText(normalLabel, x + halfW + 10, tipY - 10);
                line(x + halfW, tipY, rightColX, tipY);

                double rightColEndY;
                if (normalBranch != null && !isContinue(normalBranch) && !isBreak(normalBranch)) {
                    arrow(rightColX, tipY, rightColX, colStartY);
                    rightColEndY = renderBreakColumnChain(normalBranch, rightColX, colStartY, loop);
                } else {
                    rightColEndY = colStartY;
                }

                breakGeometry.put(node, new double[]{leftColX, rightColX, leftColEndY, rightColEndY, 1.0, tipY, colStartY});

                boolean bothBare = (leftColEndY <= colStartY + 1) && (rightColEndY <= colStartY + 1);
                double blockBottom = bothBare
                        ? (y + h)
                        : (Math.max(leftColEndY, rightColEndY) + VERTICAL_SPACING);
                node.setSize(w, blockBottom - y);
                updateMaxY(blockBottom);
                return -blockBottom;
            }
        }

        if (isBreakDecision && insideDoWhile) {
            FlowchartNode breakBranch  = trueEndsWithBreak ? node.getTrueBranch()  : node.getFalseBranch();
            FlowchartNode normalBranch = trueEndsWithBreak ? node.getFalseBranch() : node.getTrueBranch();
            String breakLabel  = trueEndsWithBreak ? "ДА"  : "НЕТ";
            String normalLabel = trueEndsWithBreak ? "НЕТ" : "ДА";

            double rightTipX  = x + halfW;
            double rightTipY  = y + halfH;
            double bottomTipX = x;
            double bottomTipY = y + h;

            labelText(breakLabel, rightTipX + 10, rightTipY - 10);

            double breakColX   = x + HORIZONTAL_SPACING;
            double colStartY   = bottomTipY + VERTICAL_SPACING;
            trackX(breakColX + PROCESS_WIDTH / 2);
            line(rightTipX, rightTipY, breakColX, rightTipY);

            double breakColEndY;
            if (isBreak(breakBranch)) {
                breakColEndY = rightTipY;
            } else {
                arrow(breakColX, rightTipY, breakColX, colStartY);
                breakColEndY = renderBreakColumnChain(breakBranch, breakColX, colStartY, null);
            }

            double exitFromY = (breakColEndY <= rightTipY + 1)
                    ? rightTipY
                    : breakColEndY;
            doWhileBreakExits.add(new double[]{breakColX, exitFromY});

            labelText(normalLabel, bottomTipX + 8, bottomTipY + 15);

            double normalEndY;
            if (normalBranch != null) {
                normalEndY = renderInlineChain(normalBranch, doWhileLoop, bottomTipX, bottomTipY + VERTICAL_SPACING, continueExits, bottomTipY);
            } else {
                normalEndY = bottomTipY;
            }

            double blockBottom = Math.max(Math.abs(normalEndY), breakColEndY);
            node.setSize(w, blockBottom - y);
            updateMaxY(blockBottom);
            return Math.abs(normalEndY);
        }

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

            labelText(breakLabel, x - halfW - 30, tipY - 10);
            line(x - halfW, tipY, daColX, tipY);

            double daColEndY;
            if (isBreak(breakBranch)) {
                daColEndY = colStartY;
            } else {
                arrow(daColX, tipY, daColX, colStartY);
                daColEndY = renderBreakColumnChain(breakBranch, daColX, colStartY, loop);
            }

            labelText(continueLabel, x + halfW + 10, tipY - 10);
            line(x + halfW, tipY, netColX, tipY);

            double netColEndY;
            boolean netIsBare = (continueBranch == null || isContinue(continueBranch) || isBreak(continueBranch));
            if (!netIsBare) {
                arrow(netColX, tipY, netColX, colStartY);
                netColEndY = renderBreakColumnChain(continueBranch, netColX, colStartY, loop);
            } else {
                netColEndY = colStartY;
            }

            double netBareFlag = netIsBare ? 1.0 : 0.0;
            boolean breakIsReturn = chainEndsWithReturn(breakBranch);
            double returnBreakFlag = breakIsReturn ? 1.0 : 0.0;

            breakGeometry.put(node, new double[]{
                    daColX, netColX, daColEndY, netColEndY,
                    0.0, tipY, colStartY, netBareFlag,
                    0.0,
                    returnBreakFlag
            });

            double blockBottom = Math.max(daColEndY, netColEndY) + VERTICAL_SPACING;
            node.setSize(w, blockBottom - y);
            updateMaxY(blockBottom);
            return -blockBottom;
        }

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
        return renderBreakColumnChain(start, x, startY, null, null);
    }

    private double renderBreakColumnChain(FlowchartNode start, double x, double startY, LoopStartNode loop) {
        return renderBreakColumnChain(start, x, startY, loop, null);
    }

    private double renderBreakColumnChain(FlowchartNode start, double x, double startY, LoopStartNode loop, FlowchartNode stopAt) {
        double currentY = startY;
        FlowchartNode cur = start;

        while (cur != null) {
            if (cur instanceof ConnectorNode) break;
            if (cur instanceof LoopEndNode)   break;
            if (cur instanceof TerminalNode)  break;
            if (stopAt != null && cur == stopAt) break;
            if (loop != null && cur == loop.getExitNode()) break;

            if (cur instanceof ProcessNode) {
                if (rendered.contains(cur)) break;
                rendered.add(cur);
                cur.setPosition(x, currentY);
                renderProcess(cur, x, currentY);
                trackX(x + PROCESS_WIDTH / 2);
                updateMaxY(currentY + PROCESS_HEIGHT);

                double nodeBottom = currentY + PROCESS_HEIGHT;

                FlowchartNode next = null;
                for (FlowchartNode n : cur.getNext()) {
                    if (n instanceof ConnectorNode)              break;
                    if (n instanceof LoopEndNode)                break;
                    if (n instanceof TerminalNode)               break;
                    if (stopAt != null && n == stopAt)           break;
                    if (loop != null && n == loop.getExitNode()) break;
                    next = n;
                    break;
                }

                if (next != null && !rendered.contains(next)) {
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

    private double renderTailBlocks(List<FlowchartNode> tail, double x, double topY, boolean firstHasArrow) {
        double currentY = topY;
        boolean first = true;

        for (FlowchartNode node : tail) {
            if (!(node instanceof ProcessNode)) break;
            rendered.add(node);

            if (!first) {
                double nextY = currentY + VERTICAL_SPACING;
                arrow(x, currentY, x, nextY - 5);
                currentY = nextY;
            } else if (firstHasArrow) {
                double blockY = currentY + VERTICAL_SPACING;
                arrow(x, currentY, x, blockY - 5);
                currentY = blockY;
            }

            node.setPosition(x, currentY);
            renderProcess(node, x, currentY);
            trackX(x + PROCESS_WIDTH / 2);
            updateMaxY(currentY + PROCESS_HEIGHT);

            currentY = currentY + PROCESS_HEIGHT;
            first = false;
        }

        lastBodyBlockX = x;
        lastBodyBlockBottomY = currentY;
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
                    if (next instanceof ConnectorNode conn && "break".equals(conn.getLabel())) {
                        double rightEdgeX = x + PROCESS_WIDTH / 2;
                        double midBlockY  = currentY + PROCESS_HEIGHT / 2;
                        doWhileBreakExits.add(new double[]{rightEdgeX, midBlockY});
                        rendered.add(next);
                        return -nodeBottom;
                    }
                    if (next instanceof ConnectorNode conn && "return".equals(conn.getLabel())) {
                        double rightEdgeX = x + PROCESS_WIDTH / 2;
                        doWhileBreakExits.add(new double[]{rightEdgeX, nodeBottom});
                        rendered.add(conn);
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
        return node instanceof ConnectorNode c &&
                ("break".equals(c.getLabel()) || "return".equals(c.getLabel()));
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

    private boolean chainEndsWithReturn(FlowchartNode node) {
        if (node == null) return false;
        if (node instanceof ConnectorNode c && "return".equals(c.getLabel())) return true;
        if (node instanceof ProcessNode) {
            for (FlowchartNode next : node.getNext()) {
                if (chainEndsWithReturn(next)) return true;
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
                endArrowAlreadyDrawn = false;
                return;
            }

            if (next instanceof ConnectorNode conn && "return".equals(conn.getLabel())) {
                for (FlowchartNode returnNext : conn.getNext()) {
                    if (returnNext instanceof TerminalNode t && !t.isStart()) {
                        endNode = t;
                        endArrowFromX = x;
                        endArrowFromY = prevBottom;
                        endArrowAlreadyDrawn = false;
                        return;
                    }
                }
                rendered.add(conn);
                return;
            }

            double nextY = prevBottom + VERTICAL_SPACING;
            arrow(x, prevBottom, x, nextY - 5);
            renderNode(next, x, nextY, stopBefore);
        }
    }

    private void renderDoWhile(DoWhileNode node, double x, double y, FlowchartNode stopBefore) {
        double arrowMidY = y - VERTICAL_SPACING / 2;

        List<double[]> continueExits = new ArrayList<>();

        double savedEndArrowFromX = endArrowFromX;
        double savedEndArrowFromY = endArrowFromY;

        double bodyEndY = renderDoWhileBody(node.getLoopBody(), node, x, y, continueExits);
        if (bodyEndY < 0) bodyEndY = -bodyEndY;

        endArrowFromX = savedEndArrowFromX;
        endArrowFromY = savedEndArrowFromY;

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

        double dW = DECISION_WIDTH;
        double dH = diamondHeight(node.getLabel());
        double halfW = dW / 2, halfH = dH / 2;
        node.setPosition(x, dY);
        beginNodeGroup(node);
        drawDiamondWithText(x, dY, dW, dH, node.getLabel());
        endNodeGroup();
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
                if (!doWhileBreakExits.isEmpty()) {
                    double breakJoinY = dBottom + VERTICAL_SPACING / 2;
                    for (double[] exit : doWhileBreakExits) {
                        line(exit[0], exit[1], exit[0], breakJoinY);
                        arrow(exit[0], breakJoinY, x + 1, breakJoinY);
                    }
                    doWhileBreakExits.clear();
                }
            } else {
                double exitY = dBottom + VERTICAL_SPACING;
                arrow(x, dBottom, x, exitY - 5);
                rendered.add(exitNode);
                exitNode.setPosition(x, exitY);
                renderProcess(exitNode, x, exitY);
                exitNode.setSize(PROCESS_WIDTH, PROCESS_HEIGHT);
                updateMaxY(exitY + PROCESS_HEIGHT);
                double exitBottom = exitY + PROCESS_HEIGHT;

                if (!doWhileBreakExits.isEmpty()) {
                    double breakJoinY = exitBottom + VERTICAL_SPACING / 2;
                    for (double[] exit : doWhileBreakExits) {
                        line(exit[0], exit[1], exit[0], breakJoinY);
                        arrow(exit[0], breakJoinY, x + 1, breakJoinY);
                    }
                    doWhileBreakExits.clear();
                }

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
            if (!doWhileBreakExits.isEmpty()) {
                double breakJoinY = dBottom + VERTICAL_SPACING / 2;
                for (double[] exit : doWhileBreakExits) {
                    line(exit[0], exit[1], exit[0], breakJoinY);
                    arrow(exit[0], breakJoinY, x + 1, breakJoinY);
                }
                doWhileBreakExits.clear();
            }
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
            if (cur instanceof ConnectorNode conn && "break".equals(conn.getLabel())) {
                doWhileBreakExits.add(new double[]{x + PROCESS_WIDTH / 2, currentY});
                rendered.add(cur);
                return -currentY;
            }
            if (cur instanceof ConnectorNode conn && "return".equals(conn.getLabel())) {
                doWhileBreakExits.add(new double[]{x + PROCESS_WIDTH / 2, currentY});
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
                    if (next instanceof ConnectorNode conn && ("break".equals(conn.getLabel()) || "return".equals(conn.getLabel()))) {
                        double rightEdgeX = x + PROCESS_WIDTH / 2;
                        double midBlockY  = currentY + PROCESS_HEIGHT / 2;
                        doWhileBreakExits.add(new double[]{rightEdgeX, midBlockY});
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
                boolean trueIsBreak     = chainEndsWithBreak(decNode.getTrueBranch());
                boolean falseIsBreak    = chainEndsWithBreak(decNode.getFalseBranch());
                boolean trueIsReturn    = chainEndsWithReturn(decNode.getTrueBranch());
                boolean falseIsReturn   = chainEndsWithReturn(decNode.getFalseBranch());
                boolean hasContinueBranch = trueIsContinue || falseIsContinue;
                boolean hasBreakBranch    = trueIsBreak    || falseIsBreak || trueIsReturn || falseIsReturn;

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
                if (hasContinueBranch || hasBreakBranch) {
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

            } else if (cur instanceof LoopStartNode innerLoop) {
                rendered.add(cur);

                double savedEAX = endArrowFromX;
                double savedEAY = endArrowFromY;

                renderLoop(innerLoop, x, currentY, null);

                endArrowFromX = savedEAX;
                endArrowFromY = savedEAY;

                double innerBottom = currentY + innerLoop.getHeight();
                updateMaxY(innerBottom);

                FlowchartNode next = null;
                if (innerLoop.getExitNode() != null && !rendered.contains(innerLoop.getExitNode())) {
                    for (FlowchartNode n : innerLoop.getExitNode().getNext()) {
                        if (n instanceof LoopEndNode)  continue;
                        if (n instanceof TerminalNode) continue;
                        if (n == loop.getExitNode())   continue;
                        next = n;
                        break;
                    }
                } else {
                    for (FlowchartNode n : cur.getNext()) {
                        if (n instanceof LoopEndNode)  continue;
                        if (n instanceof TerminalNode) continue;
                        if (n == loop.getExitNode())   continue;
                        next = n;
                        break;
                    }
                }

                if (next != null && !rendered.contains(next)) {
                    double nextY = innerBottom + VERTICAL_SPACING;
                    arrow(x, innerBottom, x, nextY - 5);
                    currentY = nextY;
                    cur = next;
                } else {
                    return innerBottom;
                }

            } else if (cur instanceof DoWhileNode innerDoWhile) {
                rendered.add(cur);

                double savedEAX = endArrowFromX;
                double savedEAY = endArrowFromY;

                renderDoWhile(innerDoWhile, x, currentY, null);

                endArrowFromX = savedEAX;
                endArrowFromY = savedEAY;

                double innerBottom = currentY + innerDoWhile.getHeight();
                updateMaxY(innerBottom);

                FlowchartNode next = null;
                if (innerDoWhile.getExitNode() != null && !rendered.contains(innerDoWhile.getExitNode())) {
                    for (FlowchartNode n : innerDoWhile.getExitNode().getNext()) {
                        if (n instanceof LoopEndNode)  continue;
                        if (n instanceof TerminalNode) continue;
                        if (n == loop.getExitNode())   continue;
                        next = n;
                        break;
                    }
                } else {
                    for (FlowchartNode n : cur.getNext()) {
                        if (n instanceof LoopEndNode)  continue;
                        if (n instanceof TerminalNode) continue;
                        if (n == loop.getExitNode())   continue;
                        next = n;
                        break;
                    }
                }

                if (next != null && !rendered.contains(next)) {
                    double nextY = innerBottom + VERTICAL_SPACING;
                    arrow(x, innerBottom, x, nextY - 5);
                    currentY = nextY;
                    cur = next;
                } else {
                    return innerBottom;
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

    private static final double CHAR_WIDTH_PX = 7.2;
    private static final double LINE_HEIGHT   = 17.0;

    private List<String> wrapText(String txt, double maxWidth, double fontSize) {
        double scale = fontSize / 13.0;
        double charW  = CHAR_WIDTH_PX * scale;
        int maxChars  = Math.max(1, (int) (maxWidth / charW));

        List<String> lines = new ArrayList<>();
        if (txt == null || txt.isEmpty()) { lines.add(""); return lines; }

        String[] words = txt.split(" ");
        StringBuilder cur = new StringBuilder();
        for (String word : words) {
            if (cur.length() == 0) {
                cur.append(word);
            } else if (cur.length() + 1 + word.length() <= maxChars) {
                cur.append(' ').append(word);
            } else {
                lines.add(cur.toString());
                cur = new StringBuilder(word);
            }
        }
        if (cur.length() > 0) lines.add(cur.toString());
        return lines;
    }

    private double diamondHeight(String label) {
        double usableW = DECISION_WIDTH * 0.55;
        List<String> lines = wrapText(label, usableW, 13);
        double needed = lines.size() * LINE_HEIGHT + 30;
        return Math.max(DECISION_HEIGHT, needed);
    }

    private void drawDiamondWithText(double x, double y, double w, double h, String label) {
        drawDiamond(x, y, w, h);
        double halfH = h / 2;
        double usableW = w * 0.55;
        List<String> lines = wrapText(label, usableW, 13);
        double textStartY = y + halfH - (lines.size() - 1) * LINE_HEIGHT / 2.0;
        for (int i = 0; i < lines.size(); i++) {
            svg.append(String.format(Locale.US,
                    "<text class=\"text\" x=\"%.1f\" y=\"%.1f\">%s</text>\n",
                    x, textStartY + i * LINE_HEIGHT, escapeXml(lines.get(i))));
        }
    }

    private void textWrapped(String txt, double cx, double cy, double maxWidth, double fontSize) {
        List<String> lines = wrapText(txt, maxWidth, fontSize);
        double startY = cy - (lines.size() - 1) * LINE_HEIGHT / 2.0;
        for (int i = 0; i < lines.size(); i++) {
            svg.append(String.format(Locale.US,
                    "<text class=\"text\" x=\"%.1f\" y=\"%.1f\" font-size=\"%.0f\">%s</text>\n",
                    cx, startY + i * LINE_HEIGHT, fontSize, escapeXml(lines.get(i))));
        }
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