package com.flowchart.model;

import com.fasterxml.jackson.annotation.JsonIdentityInfo;
import com.fasterxml.jackson.annotation.ObjectIdGenerators;
import flowchart.model.Location;

import java.util.ArrayList;
import java.util.List;
import java.util.UUID;

/**
 * Базовый класс для узлов блок-схемы по ГОСТ 19.701-90
 */
@JsonIdentityInfo(generator = ObjectIdGenerators.PropertyGenerator.class, property = "id")
public abstract class FlowchartNode {
    protected String id;
    protected String label;
    protected NodeType type;
    protected List<FlowchartNode> next;
    protected Location astLocation; // Связь с узлом AST для трассировки

    // Координаты для отрисовки (вычисляются LayoutEngine)
    protected double x;
    protected double y;
    protected double width;
    protected double height;

    public enum NodeType {
        TERMINAL,      // Терминатор (начало/конец) - скруглённый прямоугольник
        PROCESS,       // Процесс - прямоугольник
        DECISION,      // Решение - ромб
        INPUT_OUTPUT,  // Ввод/вывод - параллелограмм
        PREPARATION,   // Предопределённый процесс - прямоугольник с вертикальными полосами
        LOOP_START,    // Начало цикла - шестиугольник
        LOOP_END,      // Конец цикла - шестиугольник
        CONNECTOR      // Соединитель - круг
    }

    public FlowchartNode(NodeType type, String label) {
        this.id = UUID.randomUUID().toString();
        this.type = type;
        this.label = label;
        this.next = new ArrayList<>();
    }

    public void addNext(FlowchartNode node) {
        this.next.add(node);
    }

    // Getters/Setters
    public String getId() { return id; }
    public String getLabel() { return label; }
    public NodeType getType() { return type; }
    public List<FlowchartNode> getNext() { return next; }
    public Location getAstLocation() { return astLocation; }
    public void setAstLocation(Location loc) { this.astLocation = loc; }

    public double getX() { return x; }
    public double getY() { return y; }
    public double getWidth() { return width; }
    public double getHeight() { return height; }

    public void setPosition(double x, double y) {
        this.x = x;
        this.y = y;
    }

    public void setSize(double width, double height) {
        this.width = width;
        this.height = height;
    }
}
