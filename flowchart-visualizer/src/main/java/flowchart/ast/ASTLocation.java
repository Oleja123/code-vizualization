package flowchart.ast;

import com.fasterxml.jackson.annotation.JsonProperty; /**
 * Локация узла в исходном коде
 */
public class ASTLocation {
    @JsonProperty("line")
    private int line;
    
    @JsonProperty("column")
    private int column;
    
    @JsonProperty("endLine")
    private int endLine;
    
    @JsonProperty("endColumn")
    private int endColumn;
    
    public int getLine() { return line; }
    public int getColumn() { return column; }
    public int getEndLine() { return endLine; }
    public int getEndColumn() { return endColumn; }
}
