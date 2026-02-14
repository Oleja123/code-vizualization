package flowchart.model;

/**
 * Соединитель - для упрощения сложных схем
 * ГОСТ: круг
 */
public class ConnectorNode extends FlowchartNode {
    public ConnectorNode(String label) {
        super(NodeType.CONNECTOR, label);
    }
}
