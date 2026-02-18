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

    private static final double VERTICAL_SPACING   = 80;
    private static final double HORIZONTAL_SPACING = 260;

    private StringBuilder svg;
    private Set<FlowchartNode> rendered;
    private double maxY = 0;
    private double minX = Double.MAX_VALUE;
    private double maxX = Double.MIN_VALUE;

    private TerminalNode forcedEnd;
    private double lastNodeX = 700; // Track X coordinate of last rendered node

    public String render(FlowchartNode start) {

        svg      = new StringBuilder();
        rendered = new HashSet<>();
        maxY = 0;
        minX = Double.MAX_VALUE;
        maxX = Double.MIN_VALUE;
        forcedEnd = null;
        lastNodeX = 700;

        // первый проход
        renderNode(start, 700, 100);

        double padding = 60;

        // принудительно рисуем конец в самом низу
        if (forcedEnd != null) {
            double endX = lastNodeX; // Используем X-координату последнего узла
            double endY = maxY + VERTICAL_SPACING;

            // Рисуем стрелку от предыдущего узла к концу
            arrow(endX, maxY, endX, endY);

            renderTerminal(forcedEnd, endX, endY);
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

    // ── track bounds ──────────────────────────────────────────

    private void trackX(double x) {
        if (x < minX) minX = x;
        if (x > maxX) maxX = x;
    }


    private void updateMaxY(double y) {
        if (y > maxY) maxY = y;
    }

    // ── dispatch ─────────────────────────────────────────────

    private void renderNode(FlowchartNode node, double x, double y) {

        if (node == null || rendered.contains(node))
            return;

        rendered.add(node);
        node.setPosition(x, y);
        lastNodeX = x; // Track last node position

        trackX(x - PROCESS_WIDTH / 2);
        trackX(x + PROCESS_WIDTH / 2);

        switch (node.getType()) {

            case TERMINAL -> {
                TerminalNode t = (TerminalNode) node;

                if (t.isStart()) {
                    renderTerminal(node, x, y);
                    updateMaxY(y + TERMINAL_HEIGHT);
                    renderLinearNext(node, x, y);
                } else {
                    forcedEnd = t; // не рисуем сейчас, отложим до конца
                }
            }

            case PROCESS -> {
                renderProcess(node, x, y);
                updateMaxY(y + PROCESS_HEIGHT);
                renderLinearNext(node, x, y);
            }

            case DECISION -> renderDecision((DecisionNode) node, x, y);
            case LOOP_START -> renderLoop((LoopStartNode) node, x, y);
            case LOOP_END  -> { }
        }
    }


    // ── TERMINAL ─────────────────────────────────────────────

    private void renderTerminal(FlowchartNode node, double x, double y) {

        double w = TERMINAL_WIDTH;
        double h = TERMINAL_HEIGHT;

        node.setSize(w, h);

        svg.append(String.format(Locale.US,
                "<ellipse class=\"shape\" cx=\"%.1f\" cy=\"%.1f\" rx=\"%.1f\" ry=\"%.1f\"/>\n",
                x, y + h / 2, w / 2, h / 2));

        text(node.getLabel(), x, y + h / 2);

        trackX(x - w / 2);
        trackX(x + w / 2);
    }


    // ── PROCESS ──────────────────────────────────────────────
    private void renderProcess(FlowchartNode node, double x, double y) {

        node.setSize(PROCESS_WIDTH, PROCESS_HEIGHT);

        svg.append(String.format(Locale.US,
                "<rect class=\"shape\" x=\"%.1f\" y=\"%.1f\" width=\"%.1f\" height=\"%.1f\"/>\n",
                x - PROCESS_WIDTH / 2, y, PROCESS_WIDTH, PROCESS_HEIGHT));

        text(node.getLabel(), x, y + PROCESS_HEIGHT / 2);
    }



    // ── DECISION (if / else) ─────────────────────────────────
    /**
     * Схема:
     *
     *          ┌──────◇──────┐
     *       ДА ↓             → НЕТ
     *      [then]          [else]   ← если нет else — прямая линия вправо вниз
     *          └─────┬───────┘
     *                ↓  (точка слияния)
     *           [следующий узел]
     */
    private void renderDecision(DecisionNode node, double x, double y) {
        double w     = DECISION_WIDTH;
        double h     = DECISION_HEIGHT;
        double halfW = w / 2;
        double halfH = h / 2;
        node.setSize(w, h);

        // ромб
        String points = String.format(Locale.US,
                "%.1f,%.1f %.1f,%.1f %.1f,%.1f %.1f,%.1f",
                x, y,
                x + halfW, y + halfH,
                x, y + h,
                x - halfW, y + halfH);
        svg.append(String.format(Locale.US,
                "<polygon class='shape' points='%s'/>\n", points));
        text(node.getLabel(), x, y + halfH);

        double branchY = y + h + VERTICAL_SPACING;
        double leftX   = x - HORIZONTAL_SPACING;
        double rightX  = x + HORIZONTAL_SPACING;

        trackX(leftX  - PROCESS_WIDTH / 2);
        trackX(rightX + PROCESS_WIDTH / 2);

        double leftBottom  = branchY;
        double rightBottom = branchY;

        // ── TRUE (ДА) — идёт влево вниз ──
        if (node.getTrueBranch() != null) {
            line(x - halfW, y + halfH, leftX, y + halfH);
            line(leftX, y + halfH, leftX, branchY - 5);
            arrow(leftX, branchY - 5, leftX, branchY);
            labelText("ДА", x - halfW - 30, y + halfH - 10);
            renderNode(node.getTrueBranch(), leftX, branchY);
            leftBottom = branchY + node.getTrueBranch().getHeight();
        }

        // ── FALSE (НЕТ) — идёт вправо вниз ──
        if (node.getFalseBranch() != null) {
            // есть явная ветка else
            line(x + halfW, y + halfH, rightX, y + halfH);
            line(rightX, y + halfH, rightX, branchY - 5);
            arrow(rightX, branchY - 5, rightX, branchY);
            labelText("НЕТ", x + halfW + 10, y + halfH - 10);
            renderNode(node.getFalseBranch(), rightX, branchY);
            rightBottom = branchY + node.getFalseBranch().getHeight();
        } else {
            // НЕТ ветки else — рисуем пустую линию вправо и вниз до точки слияния
            // rightBottom будет выровнен с leftBottom после вычисления
            rightBottom = leftBottom; // временно; скорректируем ниже
        }

        // ── точка слияния ──
        double mergeY = Math.max(leftBottom, rightBottom) + VERTICAL_SPACING;

        // левая ветка → слияние
        if (node.getTrueBranch() != null) {
            line(leftX, leftBottom, leftX, mergeY);
            line(leftX, mergeY, x, mergeY);
        }

        // правая ветка → слияние
        if (node.getFalseBranch() != null) {
            line(rightX, rightBottom, rightX, mergeY);
            line(rightX, mergeY, x, mergeY);
        } else {
            // if без else: стрелка НЕТ идёт по правой стороне вниз и соединяется
            double noElseRightX = x + HORIZONTAL_SPACING;
            // горизонталь от правого угла ромба
            line(x + halfW, y + halfH, noElseRightX, y + halfH);
            labelText("НЕТ", x + halfW + 10, y + halfH - 10);
            // вертикаль вниз до уровня слияния
            line(noElseRightX, y + halfH, noElseRightX, mergeY);
            // горизонталь к центру
            line(noElseRightX, mergeY, x, mergeY);
            trackX(noElseRightX + 5);
        }

        // обновляем высоту узла
        node.setSize(w, mergeY - y);
        updateMaxY(mergeY);

        // ── следующий узел после слияния ──
        if (!node.getNext().isEmpty()) {
            double nextY = mergeY + VERTICAL_SPACING;
            arrow(x, mergeY, x, nextY - 5);
            // маленький "пенёк" перед стрелкой уже включён
            for (FlowchartNode next : node.getNext()) {
                renderNode(next, x, nextY);
            }
        }
    }

    // ── LOOP ─────────────────────────────────────────────────
    /**
     * Новая схема (как if):
     *
     *       ДА ←─────◇─────→ НЕТ
     *            [тело]    [выход]
     *               ↓
     *            [обратно к ромбу]
     */
    private void renderLoop(LoopStartNode node, double x, double y) {
        double w     = DECISION_WIDTH;
        double h     = DECISION_HEIGHT;
        double halfW = w / 2;
        double halfH = h / 2;
        node.setSize(w, h);

        // ромб
        String points = String.format(Locale.US,
                "%.1f,%.1f %.1f,%.1f %.1f,%.1f %.1f,%.1f",
                x, y,
                x + halfW, y + halfH,
                x, y + h,
                x - halfW, y + halfH);
        svg.append(String.format(Locale.US,
                "<polygon class='shape' points='%s'/>\n", points));
        text(node.getLabel(), x, y + halfH);

        double branchY = y + h + VERTICAL_SPACING;
        double leftX   = x - HORIZONTAL_SPACING;
        double rightX  = x + HORIZONTAL_SPACING;

        trackX(leftX  - PROCESS_WIDTH / 2);
        trackX(rightX + PROCESS_WIDTH / 2);

        // ДА — влево и вниз (тело цикла)
        line(x - halfW, y + halfH, leftX, y + halfH);
        line(leftX, y + halfH, leftX, branchY - 5);
        arrow(leftX, branchY - 5, leftX, branchY);
        labelText("ДА", x - halfW - 30, y + halfH - 10);

        if (node.getLoopBody() != null) {
            renderLoopBodyChain(node.getLoopBody(), leftX, branchY);
        }

        FlowchartNode lastBody = findLastBodyNode(node.getLoopBody());
        double bodyEndY = (lastBody != null)
                ? lastBody.getY() + lastBody.getHeight()
                : branchY;

        // возвратная стрелка: от конца тела вверх обратно к ромбу
        double returnTopY = y - VERTICAL_SPACING / 2;
        double returnLeftX = leftX - PROCESS_WIDTH - 20; // значительно левее, чтобы не пересекать блоки

        line(leftX, bodyEndY, leftX, bodyEndY + 20);
        line(leftX, bodyEndY + 20, returnLeftX, bodyEndY + 20);
        line(returnLeftX, bodyEndY + 20, returnLeftX, returnTopY);
        arrow(returnLeftX, returnTopY, x, returnTopY);

        trackX(returnLeftX - 5);
        updateMaxY(bodyEndY + 20);

        // НЕТ — вправо и вниз (выход из цикла)
        line(x + halfW, y + halfH, rightX, y + halfH);
        labelText("НЕТ", x + halfW + 10, y + halfH - 10);
        line(rightX, y + halfH, rightX, branchY - 5);
        arrow(rightX, branchY - 5, rightX, branchY);

        if (node.getExitNode() != null) {
            // Просто отрисовываем exitNode нормально
            renderNode(node.getExitNode(), rightX, branchY);
        }
    }

    private void renderLoopBodyChain(FlowchartNode node, double x, double y) {
        if (node == null || rendered.contains(node)) return;

        FlowchartNode current  = node;
        double        currentY = y;

        while (current != null && !rendered.contains(current)) {
            rendered.add(current);
            current.setPosition(x, currentY);
            lastNodeX = x; // Track position

            if (current instanceof LoopEndNode) break;

            if (current instanceof ProcessNode) {
                renderProcess(current, x, currentY);
                updateMaxY(currentY + PROCESS_HEIGHT);

                List<FlowchartNode> next = current.getNext();
                if (!next.isEmpty() && !(next.get(0) instanceof LoopEndNode)) {
                    double nextY = currentY + PROCESS_HEIGHT + VERTICAL_SPACING;
                    line(x, currentY + PROCESS_HEIGHT, x, nextY - 5);
                    arrow(x, nextY - 5, x, nextY);
                    current  = next.get(0);
                    currentY = nextY;
                } else {
                    break;
                }
            } else if (current instanceof DecisionNode) {
                renderDecision((DecisionNode) current, x, currentY);
                break;
            } else {
                break;
            }
        }
    }

    private FlowchartNode findLastBodyNode(FlowchartNode start) {
        if (start == null) return null;
        FlowchartNode     current = start;
        Set<FlowchartNode> visited = new HashSet<>();
        while (current != null && !visited.contains(current)) {
            visited.add(current);
            if (current instanceof LoopEndNode) return current;
            List<FlowchartNode> next = current.getNext();
            if (next.isEmpty()) return current;
            for (FlowchartNode n : next) {
                if (n instanceof LoopEndNode) return current;
            }
            current = next.get(0);
        }
        return current;
    }

    // ── LINEAR CHAIN ─────────────────────────────────────────

    private void renderLinearNext(FlowchartNode node, double x, double y) {

        if (node.getNext().isEmpty())
            return;

        double prevBottom = y + node.getHeight();

        for (FlowchartNode next : node.getNext()) {

            double nextY = prevBottom + VERTICAL_SPACING;

            line(x, prevBottom, x, nextY);

            if (!(next instanceof TerminalNode)) {
                arrow(x, nextY - 5, x, nextY);
            }

            renderNode(next, x, nextY);
        }
    }

    // ── PRIMITIVES ───────────────────────────────────────────
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

    /** Маленькая подпись (ДА / НЕТ) */
    private void labelText(String txt, double x, double y) {
        svg.append(String.format(Locale.US,
                "<text class=\"label\" x=\"%.1f\" y=\"%.1f\">%s</text>\n",
                x, y, escapeXml(txt)));
    }

    private String escapeXml(String s) {
        return s.replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;");
    }
}