package flowchart.model;

/**
 * Represents a do-while loop in the flowchart.
 * Body nodes are rendered first (top-down), then the condition diamond at the bottom.
 * + (true)  → back arrow left and up to body start
 * - (false) → straight down to exitNode
 */
public class DoWhileNode extends FlowchartNode {

    private FlowchartNode loopBody;
    private FlowchartNode exitNode;

    public DoWhileNode(String conditionLabel) {
        super(NodeType.DO_WHILE, conditionLabel);
    }

    public FlowchartNode getLoopBody() { return loopBody; }
    public void setLoopBody(FlowchartNode loopBody) { this.loopBody = loopBody; }

    public FlowchartNode getExitNode() { return exitNode; }
    public void setExitNode(FlowchartNode exitNode) { this.exitNode = exitNode; }
}