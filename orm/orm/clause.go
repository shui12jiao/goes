package orm

import (
	"fmt"
	"strings"
)

const (
	INSERT Keyword = iota
	DELETE
	UPDATE
	SELECT
	VALUES
	LIMIT
	WHERE
	ORDERBY
	COUNT
)

type Keyword int

type Clause struct {
	sql     map[Keyword]string
	sqlVars map[Keyword][]any
}

func (c *Clause) Set(name Keyword, vars ...any) *Clause {
	if c.sql == nil {
		c.sql = make(map[Keyword]string)
		c.sqlVars = make(map[Keyword][]any)
	}

	sql, vars := generators[name](vars...)
	c.sql[name] = sql
	c.sqlVars[name] = vars
	return c
}

func (c *Clause) Build(orders ...Keyword) (string, []any) {
	var sqls []string
	var vars []any
	for _, order := range orders {
		sql, ok := c.sql[order]
		if ok {
			sqls = append(sqls, sql)
			vars = append(vars, c.sqlVars[order]...)
		}
	}
	return strings.Join(sqls, " "), vars
}

type generator func(values ...any) (string, []any)

var generators map[Keyword]generator

func init() {
	generators = make(map[Keyword]generator)
	generators[INSERT] = _insert
	generators[DELETE] = _delete
	generators[UPDATE] = _update
	generators[SELECT] = _select
	generators[VALUES] = _values
	generators[LIMIT] = _limit
	generators[WHERE] = _where
	generators[ORDERBY] = _orderby
	generators[COUNT] = _count
}

func genBindVars(n int) string {
	var vars []string
	for i := 0; i < n; i++ {
		vars = append(vars, "?")
	}
	return strings.Join(vars, ", ")
}

// INSERT INTO $tableName ($fields)
func _insert(values ...any) (string, []any) {
	tableName := values[0]
	fields := strings.Join(values[1].([]string), ", ")
	return fmt.Sprintf("INSERT INTO %s (%v)", tableName, fields), []any{}
}

// VALUES ($v1), ($v2), ($v3)
func _values(values ...any) (string, []any) {
	var sql strings.Builder
	var vars []any
	sql.WriteString("VALUES ")
	bindStr := genBindVars(len(values[0].([]any)))
	for i, value := range values {
		v := value.([]any)
		sql.WriteString(fmt.Sprintf("(%v)", bindStr))
		if i != len(values)-1 {
			sql.WriteString(", ")
		}
		vars = append(vars, v...)
	}
	return sql.String(), vars
}

// DELETE FROM $tableName
func _delete(values ...any) (string, []any) {
	tableName := values[0]
	return fmt.Sprintf("DELETE FROM %s", tableName), values[1:]
}

// UPDATE $tableName SET $fields
func _update(values ...any) (string, []any) {
	tableName := values[0]
	m := values[1].(map[string]any)
	var keys []string
	var vars []any
	for k, v := range m {
		keys = append(keys, fmt.Sprintf("%s=?", k))
		vars = append(vars, v)
	}
	return fmt.Sprintf("UPDATE %s SET %s", tableName, strings.Join(keys, ", ")), vars
}

// SELECT $fields FROM $tableName
func _select(values ...any) (string, []any) {
	tableName := values[0]
	fields := strings.Join(values[1].([]string), ", ")
	return fmt.Sprintf("SELECT %v FROM %s", fields, tableName), []any{}
}

// LIMIT $limit
func _limit(values ...any) (string, []any) {
	return "LIMIT ?", values
}

// WHERE $expr
func _where(values ...any) (string, []any) {
	return fmt.Sprintf("WHERE %s", values[0]), values[1:]
}

// ORDER BY $fields
func _orderby(values ...any) (string, []any) {
	return fmt.Sprintf("ORDER BY %s", values[0]), values[1:]
}

func _count(values ...any) (string, []any) {
	return _select(values[0], []string{"COUNT(*)"})
}
