package flowchart.model;

/**
 * Терминатор - начало/конец программы или функции
 * ГОСТ: скруглённый прямоугольник
 */
public class TerminalNode extends FlowchartNode {
    private boolean isStart;
    
    public TerminalNode(String label, boolean isStart) {
        super(NodeType.TERMINAL, label);
        this.isStart = isStart;
    }
    
    public boolean isStart() { return isStart; }
}
