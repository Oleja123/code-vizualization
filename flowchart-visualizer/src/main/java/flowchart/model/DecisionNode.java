package flowchart.model;

/**
 * Решение - условный оператор
 * ГОСТ: ромб
 */
public class DecisionNode extends FlowchartNode {
    private FlowchartNode trueBranch;
    private FlowchartNode falseBranch;
    
    public DecisionNode(String condition) {
        super(NodeType.DECISION, condition);
    }
    
    public void setTrueBranch(FlowchartNode node) {
        this.trueBranch = node;
        addNext(node);
    }
    
    public void setFalseBranch(FlowchartNode node) {
        this.falseBranch = node;
        addNext(node);
    }
    
    public FlowchartNode getTrueBranch() { return trueBranch; }
    public FlowchartNode getFalseBranch() { return falseBranch; }
}
