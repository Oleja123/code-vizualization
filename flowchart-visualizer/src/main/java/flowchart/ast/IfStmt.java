package flowchart.ast;

import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.List;

public class IfStmt implements Statement {
    @JsonProperty("condition")
    private Expression condition;
    
    @JsonProperty("thenBlock")
    private Statement thenBlock;
    
    @JsonProperty("elseIf")
    private List<ElseIfClause> elseIfList;
    
    @JsonProperty("elseBlock")
    private Statement elseBlock;
    
    @JsonProperty("location")
    private ASTLocation location;
    
    public Expression getCondition() { return condition; }
    public Statement getThenBlock() { return thenBlock; }
    public List<ElseIfClause> getElseIfList() { return elseIfList; }
    public Statement getElseBlock() { return elseBlock; }
    public ASTLocation getLocation() { return location; }
}
