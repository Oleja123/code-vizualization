package flowchart.builder;

import flowchart.model.*;
import java.util.*;

/**
 * Добавьте вызов GraphDebugPrinter.print(root) сразу после buildFromProgram()
 * и пришлите вывод — это покажет реальную структуру графа.
 */
public class GraphDebugPrinter {

    public static void print(FlowchartNode root) {
        System.out.println("\n=== FLOWCHART GRAPH DUMP ===");
        visit(root, "", new IdentityHashMap<>());
        System.out.println("============================\n");
    }

    private static void visit(FlowchartNode node, String indent, IdentityHashMap<FlowchartNode, Integer> seen) {
        if (node == null) { System.out.println(indent + "(null)"); return; }

        if (seen.containsKey(node)) {
            System.out.println(indent + "-> [REF #" + seen.get(node) + " " + desc(node) + "]");
            return;
        }
        int id = seen.size();
        seen.put(node, id);
        System.out.println(indent + "#" + id + " " + desc(node));

        if (node instanceof LoopStartNode lsn) {
            System.out.println(indent + "  [BODY]:");
            visit(lsn.getLoopBody(), indent + "    ", seen);
            System.out.println(indent + "  [EXIT]:");
            visit(lsn.getExitNode(), indent + "    ", seen);
            return;
        }

        if (node instanceof DecisionNode dn) {
            System.out.println(indent + "  [TRUE]:");
            visit(dn.getTrueBranch(), indent + "    ", seen);
            System.out.println(indent + "  [FALSE]:");
            visit(dn.getFalseBranch(), indent + "    ", seen);
            System.out.println(indent + "  [NEXT(" + dn.getNext().size() + ")]:");
            for (FlowchartNode n : dn.getNext()) visit(n, indent + "    ", seen);
            return;
        }

        System.out.println(indent + "  [NEXT(" + node.getNext().size() + ")]:");
        for (FlowchartNode n : node.getNext()) visit(n, indent + "    ", seen);
    }

    private static String desc(FlowchartNode n) {
        return n.getType() + " \"" + n.getLabel() + "\" @" + System.identityHashCode(n);
    }
}