package orm

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTranscation(t *testing.T) {
	e, err := NewEngine("sqlite3", "test.db")
	require.NoError(t, err)
	defer e.Close()
	session := e.NewSession().Model(&User{})
	session.CreateTable()
	defer session.DropTable()

	testCases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{"rollback", func(t *testing.T) {
			res, err := e.Transaction(func(s *Session) (any, error) {
				// _, err := s.Raw("INSERT INTO User (Name) VALUES (?)", "Alice").Exec()
				_, err := s.Insert(&User{"Alice", 16})
				require.NoError(t, err)
				return nil, errors.New("Tx Error")
			})
			require.Nil(t, res)
			require.Error(t, err)
			require.Equal(t, "Tx Error", err.Error())

			var user User
			err = session.First(&user)
			require.Error(t, err)
			require.Empty(t, user)
		}},
		{"commit", func(t *testing.T) {
			res, err := e.Transaction(func(s *Session) (any, error) {
				require.NoError(t, err)
				// _, err := s.Raw("INSERT INTO User (Name) VALUES (?)", "Bob").Exec()
				_, err := s.Insert(&User{"Bob", 17})
				require.NoError(t, err)
				return nil, nil
			})
			require.Nil(t, res)
			require.NoError(t, err)

			var user User
			err = session.First(&user)
			require.NoError(t, err)
			require.Equal(t, "Bob", user.Name)
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.fn)
	}
}

func TestMigrate(t *testing.T) {
	e, err := NewEngine("sqlite3", "test.db")
	require.NoError(t, err)
	defer e.Close()
	session := e.NewSession().Model(&User{})
	session.CreateTable()
	require.Equal(t, 2, len(session.RefTable().FieldNames))

	type Cat struct {
		Name string
	}
	err = e.Migrate(&Cat{})
	require.NoError(t, err)
	require.True(t, session.Model(&Cat{}).HasTable())
	require.Equal(t, 1, len(session.RefTable().FieldNames))

	type User struct {
		Name  string
		Sex   bool
		Color string
	}
	err = e.Migrate(&User{})
	require.NoError(t, err)
	require.True(t, session.Model(&User{}).HasTable())
	require.Equal(t, 3, len(session.RefTable().FieldNames))

	session.Model(&Cat{}).DropTable()
	session.Model(&User{}).DropTable()
}
