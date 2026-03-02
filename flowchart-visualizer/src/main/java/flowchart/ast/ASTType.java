package flowchart.ast;

import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.List; /**
 * Тип данных
 */
public class ASTType {
    @JsonProperty("baseType")
    private String baseType;
    
    @JsonProperty("pointerLevel")
    private int pointerLevel;
    
    @JsonProperty("arraySizes")
    private List<Integer> arraySizes;
    
    public String getBaseType() { return baseType; }
    public int getPointerLevel() { return pointerLevel; }
    public List<Integer> getArraySizes() { return arraySizes; }
    
    @Override
    public String toString() {
        StringBuilder sb = new StringBuilder(baseType);
        for (int i = 0; i < pointerLevel; i++) sb.append("*");
        for (Integer size : arraySizes) sb.append("[").append(size).append("]");
        return sb.toString();
    }
}
