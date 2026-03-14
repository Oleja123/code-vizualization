package flowchart.ast;

import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.List; /**
 * Модель AST программы (соответствует Go структурам)
 */
public class Program {
    @JsonProperty("type")
    private String type;
    
    @JsonProperty("declarations")
    private List<Statement> declarations;
    
    @JsonProperty("location")
    private ASTLocation location;
    
    public List<Statement> getDeclarations() { return declarations; }
    public ASTLocation getLocation() { return location; }
}
