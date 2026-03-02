package flowchart.ast;

import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.List;

public class BlockStmt implements Statement {
    @JsonProperty("statements")
    private List<Statement> statements;
    
    @JsonProperty("location")
    private ASTLocation location;
    
    public List<Statement> getStatements() { return statements; }
    public ASTLocation getLocation() { return location; }
}
