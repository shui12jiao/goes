package orm

import (
	"orm/logger"
	"testing"

	"github.com/stretchr/testify/require"
)

// insert
func (u *User) BeforeInsert(s *Session) error {
	u.Age += 100
	return nil
}

func (u *User) AfterInsert(s *Session) error { //no effect database
	u.Age -= 1
	return nil
}

func TestInsertHook(t *testing.T) {
	session := session(t).EnableHook()
	session.CreateTable()
	defer session.db.Close()
	defer session.DropTable()

	affected, err := session.Insert(&User{"Insert", 18})
	require.NoError(t, err)
	require.Equal(t, int64(1), affected)

	var user User
	err = session.Raw("SELECT * FROM User WHERE Name = ?", "Insert").QueryRow().Scan(&user.Name, &user.Age)
	require.NoError(t, err)
	require.EqualValues(t, 118, user.Age)
}

// query
func (u *User) BeforeQuery(s *Session) error {
	u.Age += 100
	return nil
}

func (u *User) AfterQuery(s *Session) error {
	u.Age -= 1
	return nil
}
func TestFindHook(t *testing.T) {
	session := session(t).EnableHook()
	session.CreateTable()
	defer session.db.Close()
	defer session.DropTable()

	session.Insert(&User{"Alice", 16})
	session.Insert(&User{"Bob", 17})
	session.Insert(&User{"Cindy", 18})

	var users []User
	err := session.Find(&users)
	require.NoError(t, err)
	require.Equal(t, 3, len(users))
	require.EqualValues(t, []User{{"Alice", 115}, {"Bob", 116}, {"Cindy", 117}}, users)
}

// update
func (u *User) BeforeUpdate(s *Session) error {
	u.Age += 100
	return nil
}

func (u *User) AfterUpdate(s *Session) error {
	u.Age -= 1
	return nil
}
func TestUpdateHook(t *testing.T) {
	session := session(t).EnableHook()
	session.CreateTable()
	defer session.db.Close()
	defer session.DropTable()

	session.Insert(&User{"Insert", 18})

	affected, err := session.Where("Name = ?", "Insert").Update("Age", 19)
	require.NoError(t, err)
	require.Equal(t, int64(1), affected)

	var user User
	err = session.Raw("SELECT * FROM User WHERE Name = ?", "Insert").QueryRow().Scan(&user.Name, &user.Age)
	require.NoError(t, err)
	require.EqualValues(t, 19, user.Age)
}

// delete
func (u *User) BeforeDelete(s *Session) error {
	logger.Info("BeforeDelete")
	return nil
}

func (u *User) AfterDelete(s *Session) error {
	logger.Info("AfterDelete")
	return nil
}

func TestDeleteHook(t *testing.T) {
	session := session(t).EnableHook()
	session.CreateTable()
	defer session.db.Close()
	defer session.DropTable()

	session.Insert(&User{"Insert", 18})

	affected, err := session.Where("Name = ?", "Insert").Delete()
	require.NoError(t, err)
	require.Equal(t, int64(1), affected)

	var user User
	err = session.Raw("SELECT * FROM User WHERE Name = ?", "Insert").QueryRow().Scan(&user.Name, &user.Age)
	require.Error(t, err)
	require.Empty(t, user)
}
