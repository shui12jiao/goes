package orm

import (
	"database/sql"
	"orm/dialect"
	"testing"

	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	Name string `orm:"PRIMARY KEY"`
	Age  int    `orm:"DEFAULT(18)"`
}

func session(t *testing.T) *Session {
	t.Helper()
	db, err := sql.Open("sqlite3", "test.db")
	require.NoError(t, err)
	d, ok := dialect.GetDialect("sqlite3")
	require.True(t, ok)

	return NewSession(db, d).Model(&User{})
}

func TestTable(t *testing.T) {
	session := session(t)
	defer session.db.Close()

	err := session.CreateTable()
	require.NoError(t, err)
	exist := session.HasTable()
	require.True(t, exist)
	err = session.DropTable()
	require.NoError(t, err)
}

func TestInsert(t *testing.T) {
	session := session(t)
	session.CreateTable()
	defer session.db.Close()
	defer session.DropTable()

	affected, err := session.Insert(&User{"Insert", 18})
	require.NoError(t, err)
	require.Equal(t, int64(1), affected)
}

func TestFind(t *testing.T) {
	session := session(t)
	session.CreateTable()
	defer session.db.Close()
	defer session.DropTable()

	session.Insert(&User{"Alice", 16})
	session.Insert(&User{"Bob", 17})
	session.Insert(&User{"Cindy", 18})

	var users []User
	err := session.Find(&users)
	require.NoError(t, err)

	rows, err := session.Raw("SELECT * FROM User").QueryRows()
	require.NoError(t, err)
	defer rows.Close()
	var usersT []User
	for rows.Next() {
		var user User
		err = rows.Scan(&user.Name, &user.Age)
		require.NoError(t, err)
		usersT = append(usersT, user)
	}
	require.EqualValues(t, users, usersT)
}

func TestFirst(t *testing.T) {
	session := session(t)
	session.CreateTable()
	defer session.db.Close()
	defer session.DropTable()

	session.Insert(&User{"First", 16})
	session.Insert(&User{"Last", 17})

	var user User
	err := session.First(&user)
	require.NoError(t, err)

	var userT User
	row := session.Raw("SELECT * FROM User LIMIT 1").QueryRow()
	defer session.Raw("DELETE FROM User WHERE Name = ?", "First").Exec()
	err = row.Scan(&userT.Name, &userT.Age)
	require.NoError(t, err)

	require.EqualValues(t, userT, user)
}

func TestUpdate(t *testing.T) {
	session := session(t)
	session.CreateTable()
	defer session.db.Close()
	defer session.DropTable()

	session.Insert(&User{"Update", 16})

	affected, err := session.Where("Name = ?", "Update").Update("Age", 18)
	require.NoError(t, err)
	require.Equal(t, int64(1), affected)

	var user User
	err = session.Raw("SELECT * FROM User WHERE Name = ?", "Update").QueryRow().Scan(&user.Name, &user.Age)
	require.NoError(t, err)
	require.EqualValues(t, 18, user.Age)
}

func TestDelete(t *testing.T) {
	session := session(t)
	session.CreateTable()
	defer session.db.Close()
	defer session.DropTable()

	session.Insert(&User{"Delete", 16})
	affected, err := session.Where("Name = ?", "Delete").Delete()
	require.NoError(t, err)
	require.Equal(t, int64(1), affected)

	var user User
	err = session.Raw("SELECT * FROM User WHERE Name = ?", "Delete").QueryRow().Scan(&user.Name, &user.Age)
	require.Error(t, err)
	require.Empty(t, user)
}
