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

    private static final double VERTICAL_SPACING = 80;
    private static final double HORIZONTAL_SPACING = 250;

    private StringBuilder svg;
    private Set<FlowchartNode> rendered;
    private double maxY = 0;

    public String render(FlowchartNode start) {
        svg = new StringBuilder();
        rendered = new HashSet<>();
        maxY = 0;

        svg.append("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n");
        svg.append("<svg xmlns=\"http://www.w3.org/2000/svg\" ");
        svg.append("width=\"1400\" height=\"2000\" viewBox=\"0 0 1400 2000\">\n");
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
                updateMaxY(y + TERMINAL_HEIGHT);
                renderLinearNext(node, x, y);
            }
            case PROCESS -> {
                renderProcess(node, x, y);
                updateMaxY(y + PROCESS_HEIGHT);
                renderLinearNext(node, x, y);
            }
            case DECISION -> {
                renderDecision((DecisionNode) node, x, y);
            }
            case LOOP_START -> {
                renderLoop((LoopStartNode) node, x, y);
            }
            case LOOP_END -> {
                // Не рендерится отдельно
            }
        }
    }

    private void updateMaxY(double y) {
        if (y > maxY) maxY = y;
    }

    /**
     * Компактная отрисовка цикла как на скрине:
     * - Ромб
     * - ДА вниз из нижнего угла
     * - НЕТ вправо из правого угла
     * - Стрелка назад к левому углу
     */
    private void renderLoop(LoopStartNode node, double x, double y) {

        double w = DECISION_WIDTH;
        double h = DECISION_HEIGHT;
        double halfW = w / 2;
        double halfH = h / 2;

        node.setSize(w, h);

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

        // ===== ДА вниз =====
        double bodyY = y + h + VERTICAL_SPACING;

        line(x, y + h, x, bodyY - 5);
        arrow(x, bodyY - 5, x, bodyY);
        text("ДА", x + 20, y + h + 20);

        // Рисуем тело
        if (node.getLoopBody() != null) {
            renderLoopBodyChain(node.getLoopBody(), x, bodyY);
        }

        FlowchartNode lastBody = findLastBodyNode(node.getLoopBody());
        double bodyEndY = lastBody != null
                ? lastBody.getY() + lastBody.getHeight()
                : bodyY;

        // =====================================================
        // 🔁 ВОЗВРАТНАЯ СТРЕЛКА (аккуратная, выше ромба)
        // =====================================================

        double returnJoinY = y - VERTICAL_SPACING / 2;   // чуть выше ромба
        double leftX = x - HORIZONTAL_SPACING;

        // вниз немного от тела
        line(x, bodyEndY, x, bodyEndY + 20);

        // влево
        line(x, bodyEndY + 20, leftX, bodyEndY + 20);

        // вверх к точке соединения
        line(leftX, bodyEndY + 20, leftX, returnJoinY);

        // стрелка вправо к середине входящей линии
        arrow(leftX, returnJoinY, x, returnJoinY);

        updateMaxY(bodyEndY + 20);

        // =====================================================
        // ❌ НЕТ вправо (как у if)
        // =====================================================

        double rightX = x + HORIZONTAL_SPACING;
        double exitY = bodyEndY + VERTICAL_SPACING;

        line(x + halfW, y + halfH, rightX, y + halfH);
        text("НЕТ", x + halfW + 30, y + halfH - 10);

        line(rightX, y + halfH, rightX, exitY);
        line(rightX, exitY, x, exitY);

        updateMaxY(exitY);

        // После цикла
        if (node.getExitNode() != null) {
            double nextY = exitY + VERTICAL_SPACING;

            line(x, exitY, x, nextY - 5);
            arrow(x, nextY - 5, x, nextY);

            renderNode(node.getExitNode(), x, nextY);
        }
    }

    /**
     * Рендерим цепочку узлов в теле цикла
     */
    private void renderLoopBodyChain(FlowchartNode node, double x, double y) {
        if (node == null || rendered.contains(node)) return;

        FlowchartNode current = node;
        double currentY = y;

        while (current != null && !rendered.contains(current)) {
            rendered.add(current);
            current.setPosition(x, currentY);

            if (current instanceof LoopEndNode) {
                break;
            }

            if (current instanceof ProcessNode) {
                renderProcess(current, x, currentY);
                updateMaxY(currentY + PROCESS_HEIGHT);

                List<FlowchartNode> next = current.getNext();
                if (!next.isEmpty() && !(next.get(0) instanceof LoopEndNode)) {
                    double nextY = currentY + PROCESS_HEIGHT + VERTICAL_SPACING;
                    line(x, currentY + PROCESS_HEIGHT, x, nextY - 5);
                    arrow(x, nextY - 5, x, nextY);
                    current = next.get(0);
                    currentY = nextY;
                } else {
                    break;
                }
            } else if (current instanceof DecisionNode) {
                // Decision внутри цикла
                renderDecision((DecisionNode) current, x, currentY);
                break;
            } else {
                break;
            }
        }
    }

    private FlowchartNode findLastBodyNode(FlowchartNode start) {
        if (start == null) return null;

        FlowchartNode current = start;
        Set<FlowchartNode> visited = new HashSet<>();

        while (current != null && !visited.contains(current)) {
            visited.add(current);

            if (current instanceof LoopEndNode) {
                return current;
            }

            List<FlowchartNode> next = current.getNext();
            if (next.isEmpty()) {
                return current;
            }

            for (FlowchartNode n : next) {
                if (n instanceof LoopEndNode) {
                    return current;
                }
            }

            current = next.get(0);
        }

        return current;
    }

    private void renderDecision(DecisionNode node, double x, double y) {
        node.setSize(DECISION_WIDTH, DECISION_HEIGHT);

        double w = DECISION_WIDTH;
        double h = DECISION_HEIGHT;
        double halfW = w / 2;
        double halfH = h / 2;

        // Ромб
        String points = String.format(Locale.US,
                "%.1f,%.1f %.1f,%.1f %.1f,%.1f %.1f,%.1f",
                x, y,
                x + halfW, y + halfH,
                x, y + h,
                x - halfW, y + halfH);
        svg.append(String.format(Locale.US, "<polygon class='shape' points='%s'/>", points));
        text(node.getLabel(), x, y + h / 2);

        double branchY = y + h + VERTICAL_SPACING;
        double leftX = x - HORIZONTAL_SPACING;
        double rightX = x + HORIZONTAL_SPACING;

        double leftBottom = branchY;
        double rightBottom = branchY;

        // TRUE
        if (node.getTrueBranch() != null) {
            line(x - halfW, y + halfH, leftX, y + halfH);
            line(leftX, y + halfH, leftX, branchY - 5);
            arrow(leftX, branchY - 5, leftX, branchY);
            text("ДА", x - halfW - 30, y + halfH - 10);

            renderNode(node.getTrueBranch(), leftX, branchY);
            leftBottom = branchY + node.getTrueBranch().getHeight();
        }

        // FALSE
        if (node.getFalseBranch() != null) {
            line(x + halfW, y + halfH, rightX, y + halfH);
            line(rightX, y + halfH, rightX, branchY - 5);
            arrow(rightX, branchY - 5, rightX, branchY);
            text("НЕТ", x + halfW + 30, y + halfH - 10);

            renderNode(node.getFalseBranch(), rightX, branchY);
            rightBottom = branchY + node.getFalseBranch().getHeight();
        }

        // Слияние
        double mergeY = Math.max(leftBottom, rightBottom) + VERTICAL_SPACING;

        line(leftX, leftBottom, leftX, mergeY);
        line(leftX, mergeY, x, mergeY);
        line(rightX, rightBottom, rightX, mergeY);
        line(rightX, mergeY, x, mergeY);

        node.setSize(DECISION_WIDTH, mergeY - y);
        updateMaxY(mergeY);

        // Следующие после слияния
        if (!node.getNext().isEmpty()) {
            double nextY = mergeY + VERTICAL_SPACING;
            for (FlowchartNode next : node.getNext()) {
                line(x, mergeY, x, nextY - 5);
                arrow(x, nextY - 5, x, nextY);
                renderNode(next, x, nextY);
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

    private void renderTerminalEnd(double x, double y) {
        double w = TERMINAL_WIDTH, h = TERMINAL_HEIGHT;

        svg.append(String.format(Locale.US,
                "<ellipse class=\"shape\" cx=\"%.1f\" cy=\"%.1f\" rx=\"%.1f\" ry=\"%.1f\"/>\n",
                x, y + h / 2, w / 2, h / 2));
        text("конец", x, y + h / 2);
    }

    private void renderProcess(FlowchartNode node, double x, double y) {
        node.setSize(PROCESS_WIDTH, PROCESS_HEIGHT);
        svg.append(String.format(Locale.US,
                "<rect class=\"shape\" x=\"%.1f\" y=\"%.1f\" width=\"%.1f\" height=\"%.1f\"/>\n",
                x - PROCESS_WIDTH / 2, y, PROCESS_WIDTH, PROCESS_HEIGHT));
        text(node.getLabel(), x, y + PROCESS_HEIGHT / 2);
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