package tables

type Table struct {
	Header    []string
	Value     [][]string
	Soort     bool
	TableName string
}

func NewTable(header []string, value [][]string, soort bool, tableName string) *Table {
	return &Table{header, value, soort, tableName}
}
