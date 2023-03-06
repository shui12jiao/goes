package dialect

import "reflect"

type sqlite3 struct{}

// var _ Dialect = (*sqlite3)(nil)

func init() {
	RegisterDialect("sqlite3", &sqlite3{})
}

func (s *sqlite3) TypeMap(typ reflect.Type) string {
	switch typ.Kind() {
	case reflect.Bool:
		return "BOOLEAN"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return "INTEGER"
	case reflect.Float32, reflect.Float64:
		return "REAL"
	case reflect.String:
		return "TEXT"
	case reflect.Array, reflect.Slice:
		return "BLOB"
	case reflect.Struct:
		if _, ok := typ.FieldByName("Time"); ok {
			return "DATETIME"
		}
	default:
	}

	panic("invalid sql type " + typ.String())
}

func (s *sqlite3) TableExitSQL(tableName string) (string, []any) {
	return "SELECT name FROM sqlite_master WHERE type='table' AND name=?;", []any{tableName}
}
