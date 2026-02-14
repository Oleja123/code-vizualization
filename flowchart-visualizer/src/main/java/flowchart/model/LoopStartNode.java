package flowchart.model;

/**
 * Начало цикла (for/while)
 * ГОСТ 19.701-90: Граница цикла - трапеция с условием внутри
 * Обе части (начало и конец) имеют одинаковый идентификатор
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
