package flowchart.model;

/**
 * Начало цикла (for/while)
 * ГОСТ: шестиугольник
 */
public class LoopStartNode extends FlowchartNode {
    private FlowchartNode loopBody;
    private FlowchartNode exitNode;

    public LoopStartNode(String condition) {
        super(NodeType.LOOP_START, condition);
    }

    public void setLoopBody(FlowchartNode node) {
        this.loopBody = node;
        addNext(node);
    }

    public void setExitNode(FlowchartNode node) {
        this.exitNode = node;
        addNext(node);
    }

    public FlowchartNode getLoopBody() { return loopBody; }
    public FlowchartNode getExitNode() { return exitNode; }
}
