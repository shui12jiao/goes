package orm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClause(t *testing.T) {
	testCases := []struct {
		name     string
		do       func(expected string)
		expected string
	}{
		{
			name: "select",
			do: func(expected string) {
				c := Clause{}
				c.Set(SELECT, "User", []string{"name", "age"}).Set(WHERE, "name = ?", "Alice").Set(ORDERBY, "age DESC").Set(LIMIT, 7)
				sql, vars := c.Build(SELECT, WHERE, ORDERBY, LIMIT)
				require.Equal(t, expected, sql)
				require.Equal(t, []any{"Alice", 7}, vars)
			},
			expected: "SELECT name, age FROM User WHERE name = ? ORDER BY age DESC LIMIT ?",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.do(tc.expected)
		})
	}
}
