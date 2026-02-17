package flowchart.model;

/**
 * Конец цикла - возврат к началу
 * ГОСТ: шестиугольник
 */
public class LoopEndNode extends FlowchartNode {
    private FlowchartNode loopStart;
    
    public LoopEndNode() {
        super(NodeType.LOOP_END, "");
    }
    
    public void setLoopStart(FlowchartNode node) {
        this.loopStart = node;
        addNext(node);
    }
    
    public FlowchartNode getLoopStart() { return loopStart; }
}
