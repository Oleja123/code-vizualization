package flowchart.model;

/**
 * Позиция в исходном коде (соответствует Location из AST)
 * Используется для связывания блоков схемы с исходным кодом
 */
public class Location {
    private int line;
    private int column;
    private int endLine;
    private int endColumn;
    
    public Location(int line, int column, int endLine, int endColumn) {
        this.line = line;
        this.column = column;
        this.endLine = endLine;
        this.endColumn = endColumn;
    }
    
    // Getters
    public int getLine() { return line; }
    public int getColumn() { return column; }
    public int getEndLine() { return endLine; }
    public int getEndColumn() { return endColumn; }
    
    @Override
    public String toString() {
        return String.format("L%d:%d-%d:%d", line, column, endLine, endColumn);
    }
}
