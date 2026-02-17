package flowchart.ast;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

import java.util.List;

// Go шлёт: {"type":"IfStmt","condition":...,"thenBlock":...,"elseBlock":...}
// elseIf представлен как elseBlock с вложенным IfStmt (не отдельный список)
@JsonIgnoreProperties(ignoreUnknown = true)
public class IfStmt implements Statement {
    @JsonProperty("condition")
    private Expression condition;

    @JsonProperty("thenBlock")
    private Statement thenBlock;

    @JsonProperty("elseBlock")
    private Statement elseBlock;

    @JsonProperty("location")
    private ASTLocation location;

    public Expression getCondition() { return condition; }
    public Statement getThenBlock() { return thenBlock; }
    public Statement getElseBlock() { return elseBlock; }
    public ASTLocation getLocation() { return location; }

    // Совместимость со старым кодом — elseIfList всегда пустой,
    // цепочка else-if уже встроена в elseBlock
    public List<ElseIfClause> getElseIfList() { return null; }
}