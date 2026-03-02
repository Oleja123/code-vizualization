package flowchart.model;

public class PositionedNode {
    public FlowchartNode node;
    public int x;
    public int y;

    public PositionedNode(FlowchartNode node, int x, int y) {
        this.node = node;
        this.x = x;
        this.y = y;
    }
}
