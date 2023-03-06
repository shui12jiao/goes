package dialect

import "reflect"

var dialects = map[string]Dialect{}

type Dialect interface {
	TypeMap(typ reflect.Type) string
	TableExitSQL(tableName string) (string, []any)
}

func RegisterDialect(name string, dialect Dialect) {
	dialects[name] = dialect
}

func GetDialect(name string) (d Dialect, ok bool) {
	d, ok = dialects[name]
	return
}
