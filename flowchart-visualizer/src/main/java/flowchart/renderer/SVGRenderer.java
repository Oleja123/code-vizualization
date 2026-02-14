package flowchart.renderer;

import flowchart.model.*;
import java.util.*;

public class SVGRenderer {

    private static final double PROCESS_WIDTH = 220;
    private static final double PROCESS_HEIGHT = 70;
    private static final double TERMINAL_WIDTH = 220;
    private static final double TERMINAL_HEIGHT = 60;
    private static final double DECISION_WIDTH = 220;
    private static final double DECISION_HEIGHT = 120;

    private static final double VERTICAL_SPACING = 100;
    private static final double HORIZONTAL_SPACING = 300;

    private StringBuilder svg;
    private Set<FlowchartNode> rendered;

    public String render(FlowchartNode start) {
        svg = new StringBuilder();
        rendered = new HashSet<>();

        svg.append("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n");
        svg.append("<svg xmlns=\"http://www.w3.org/2000/svg\" ");
        svg.append("width=\"1400\" height=\"2000\" viewBox=\"0 0 1400 1000\">\n");
        svg.append("<defs>\n");
        svg.append("<marker id=\"arrow\" markerWidth=\"10\" markerHeight=\"10\" refX=\"9\" refY=\"5\" orient=\"auto\">\n");
        svg.append("<path d=\"M0,0 L10,5 L0,10 z\" fill=\"black\"/>\n");
        svg.append("</marker>\n");
        svg.append("<style>\n");
        svg.append(".shape { fill: white; stroke: black; stroke-width: 2; }\n");
        svg.append(".line { stroke: black; stroke-width: 2; fill: none; }\n");
        svg.append(".arrow { stroke: black; stroke-width: 2; fill: none; marker-end: url(#arrow); }\n");
        svg.append(".text { font-family: Arial; font-size: 13px; text-anchor: middle; dominant-baseline: middle; }\n");
        svg.append(".label { font-family: Arial; font-size: 11px; fill: #333; }\n");
        svg.append("</style>\n");
        svg.append("</defs>\n");

        renderNode(start, 700, 100);

        svg.append("</svg>");
        return svg.toString();
    }

    private void renderNode(FlowchartNode node, double x, double y) {
        if (node == null || rendered.contains(node)) return;
        rendered.add(node);
        node.setPosition(x, y);

        switch (node.getType()) {
            case TERMINAL -> {
                renderTerminal(node, x, y);
                renderLinearNext(node, x, y);
            }
            case PROCESS -> {
                renderProcess(node, x, y);
                renderLinearNext(node, x, y);
            }
            case DECISION -> {
                renderDecision((DecisionNode) node, x, y);
            }
            case LOOP_START -> {
                renderLoopStart((LoopStartNode) node, x, y);
                renderLoopBranches((LoopStartNode) node, x, y);
            }
            case LOOP_END -> {
                renderLoopEnd(node, x, y);
                renderLoop((LoopEndNode) node, x, y);
            }
        }
    }
    
    private void renderDecision(DecisionNode node, double x, double y) {

        node.setSize(DECISION_WIDTH, DECISION_HEIGHT);

        double w = DECISION_WIDTH;
        double h = DECISION_HEIGHT;

        double halfW = w / 2;
        double halfH = h / 2;

        // ===== РОМБ =====
        String points = String.format(Locale.US,
                "%.1f,%.1f %.1f,%.1f %.1f,%.1f %.1f,%.1f",
                x, y,
                x + halfW, y + halfH,
                x, y + h,
                x - halfW, y + halfH);

        svg.append(String.format(Locale.US,
                "<polygon class='shape' points='%s'/>", points));

        text(node.getLabel(), x, y + h / 2);

        double branchStartY = y + h;
        double branchY = branchStartY + VERTICAL_SPACING;

        double leftX = x - HORIZONTAL_SPACING;
        double rightX = x + HORIZONTAL_SPACING;

        double leftBottom = branchY;
        double rightBottom = branchY;

        // ===== TRUE =====
        if (node.getTrueBranch() != null) {

            line(x - halfW, y + halfH, leftX, y + halfH);
            line(leftX, y + halfH, leftX, branchY - 5);
            arrow(leftX, branchY - 5, leftX, branchY);

            text("да", x - halfW - 30, y + halfH - 10);

            renderNode(node.getTrueBranch(), leftX, branchY);
            leftBottom = branchY + node.getTrueBranch().getHeight();
        }

        // ===== FALSE =====
        if (node.getFalseBranch() != null) {

            line(x + halfW, y + halfH, rightX, y + halfH);
            line(rightX, y + halfH, rightX, branchY - 5);
            arrow(rightX, branchY - 5, rightX, branchY);

            text("нет", x + halfW + 30, y + halfH - 10);

            renderNode(node.getFalseBranch(), rightX, branchY);
            rightBottom = branchY + node.getFalseBranch().getHeight();
        }

        // ===== СВЕДЕНИЕ =====
        double mergeY = Math.max(leftBottom, rightBottom) + VERTICAL_SPACING;

        // ЛИНИИ БЕЗ СТРЕЛОК от веток к точке слияния
        line(leftX, leftBottom, leftX, mergeY);
        line(leftX, mergeY, x, mergeY);

        line(rightX, rightBottom, rightX, mergeY);
        line(rightX, mergeY, x, mergeY);

        // Стрелка ПОСЛЕ слияния вниз
        double afterMergeY = mergeY + VERTICAL_SPACING;
        line(x, mergeY, x, afterMergeY - 5);
        arrow(x, afterMergeY - 5, x, afterMergeY);

        // 🔥 КРИТИЧНО — обновляем высоту decision (включая слияние)
        node.setSize(DECISION_WIDTH, afterMergeY - y);

        // Рендерим следующие узлы ПОСЛЕ точки слияния
        if (!node.getNext().isEmpty()) {
            for (FlowchartNode next : node.getNext()) {
                renderNode(next, x, afterMergeY);
            }
        }
    }


    private void renderLinearNext(FlowchartNode node, double x, double y) {
        if (node.getNext().isEmpty()) return;

        double prevBottom = y + node.getHeight();

        for (FlowchartNode next : node.getNext()) {
            double nextY = prevBottom + VERTICAL_SPACING;

            line(x, prevBottom, x, nextY - 5);
            arrow(x, nextY - 5, x, nextY);

            renderNode(next, x, nextY);
        }
    }

    private void renderTerminal(FlowchartNode node, double x, double y) {
        double w = TERMINAL_WIDTH, h = TERMINAL_HEIGHT;
        node.setSize(w, h);

        svg.append(String.format(Locale.US,
                "<ellipse class=\"shape\" cx=\"%.1f\" cy=\"%.1f\" rx=\"%.1f\" ry=\"%.1f\"/>\n",
                x, y + h / 2, w / 2, h / 2));

        text(node.getLabel(), x, y + h / 2);
    }

    private void renderProcess(FlowchartNode node, double x, double y) {
        node.setSize(PROCESS_WIDTH, PROCESS_HEIGHT);
        svg.append(String.format(Locale.US,
                "<rect class=\"shape\" x=\"%.1f\" y=\"%.1f\" width=\"%.1f\" height=\"%.1f\"/>\n",
                x - PROCESS_WIDTH / 2, y, PROCESS_WIDTH, PROCESS_HEIGHT));
        text(node.getLabel(), x, y + PROCESS_HEIGHT / 2);
    }

    private void renderLoopStart(LoopStartNode node, double x, double y) {
        double w = 140, h = 60, cut = 20;
        node.setSize(w, h);

        double left = x - w / 2, right = x + w / 2, top = y, bottom = y + h;

        String points = String.format(Locale.US,
                "%.1f,%.1f %.1f,%.1f %.1f,%.1f %.1f,%.1f %.1f,%.1f %.1f,%.1f",
                left + cut, top,
                right - cut, top,
                right, top + cut,
                right, bottom,
                left, bottom,
                left, top + cut);

        svg.append(String.format(Locale.US, "<polygon class='shape' points='%s'/>", points));
        text(node.getLabel(), x, y + h / 2);
    }

    private void renderLoopEnd(FlowchartNode node, double x, double y) {
        double w = 140, h = 60, cut = 20;
        node.setSize(w, h);

        double left = x - w / 2, right = x + w / 2, top = y, bottom = y + h;

        String points = String.format(Locale.US,
                "%.1f,%.1f %.1f,%.1f %.1f,%.1f %.1f,%.1f %.1f,%.1f %.1f,%.1f",
                left, top,
                right, top,
                right, bottom - cut,
                right - cut, bottom,
                left + cut, bottom,
                left, bottom - cut);

        svg.append(String.format(Locale.US, "<polygon class='shape' points='%s'/>", points));
        text(node.getLabel(), x, y + h / 2);
    }

    private void renderLoopBranches(LoopStartNode node, double x, double y) {
        if (node.getLoopBody() != null) {
            double bodyY = y + DECISION_HEIGHT + VERTICAL_SPACING;
            line(x, y + DECISION_HEIGHT, x, bodyY - 5);
            arrow(x, bodyY - 5, x, bodyY);
            renderNode(node.getLoopBody(), x, bodyY);
        }
    }

    private void renderLoop(LoopEndNode node, double x, double y) {
        // Можно добавить стрелку обратно к LoopStart, если нужно
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

    private String escapeXml(String s) {
        return s.replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;");
    }

}
