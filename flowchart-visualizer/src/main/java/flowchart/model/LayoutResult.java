package flowchart.model;

import java.util.ArrayList;
import java.util.List;

public class LayoutResult {
    public List<PositionedNode> nodes = new ArrayList<>();
    public int bottomY;

    public LayoutResult(int bottomY) {
        this.bottomY = bottomY;
    }

    public void merge(LayoutResult other) {
        nodes.addAll(other.nodes);
        bottomY = Math.max(bottomY, other.bottomY);
    }
}