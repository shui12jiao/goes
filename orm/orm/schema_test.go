package orm

import (
	"orm/dialect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	type User struct {
		Name string `orm:"PRIMARY KEY"`
		Age  int    `orm:"DEFAULT(18)"`
	}
	dialect, ok := dialect.GetDialect("sqlite3")
	require.True(t, ok)
	schema := Parse(&User{Name: "Alice", Age: 16}, dialect)
	require.Equal(t, "User", schema.Name)
	require.Equal(t, 2, len(schema.Fields))
	require.Equal(t, "PRIMARY KEY", schema.GetFiled("Name").Tag)
	require.Equal(t, "DEFAULT(18)", schema.GetFiled("Age").Tag)
}

func TestRecordValues(t *testing.T) {
	type User struct {
		Name string `orm:"PRIMARY KEY"`
		Age  int    `orm:"DEFAULT(18)"`
	}
	dialect, ok := dialect.GetDialect("sqlite3")
	require.True(t, ok)
	schema := Parse(&User{}, dialect)
	fv := schema.RecordValues(&User{Name: "Alice", Age: 16})
	require.Equal(t, 2, len(fv))
	require.EqualValues(t, []any{"Alice", 16}, fv)
}
