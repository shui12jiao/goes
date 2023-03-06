package orm

import (
	"database/sql"
	"fmt"
	"orm/dialect"
	"orm/logger"
	"strings"
)

type Engine struct {
	db      *sql.DB
	dialect dialect.Dialect
}

func NewEngine(driver, source string) (e *Engine, err error) {
	db, err := sql.Open(driver, source)
	if err != nil {
		logger.Error(err)
		return
	}

	err = db.Ping()
	if err != nil {
		logger.Error(err)
		return
	}

	dialect, ok := dialect.GetDialect(driver)
	if !ok {
		logger.Errorf("dialect %s not found", driver)
		return
	}
	e = &Engine{db: db, dialect: dialect}
	logger.Info("database connected")
	return
}

func (e *Engine) Close() {
	err := e.db.Close()
	if err != nil {
		logger.Errorf("failed to close database: %v\n", err)
		return
	}
	logger.Info("database closed")
}

func (e *Engine) NewSession() *Session {
	return NewSession(e.db, e.dialect)
}

type Txfunc func(*Session) (any, error)

func (e *Engine) Transaction(f Txfunc) (result any, err error) {
	s := e.NewSession()
	if err = s.Begin(); err != nil {
		return
	}
	defer func() {
		if p := recover(); p != nil {
			_ = s.Rollback()
			panic(p)
		} else if err != nil {
			_ = s.Rollback()
		} else {
			err = s.Commit()
		}
	}()
	return f(s)
}

func (e *Engine) Migrate(value any) error {
	_, err := e.Transaction(func(s *Session) (result any, err error) {
		if !s.Model(value).HasTable() {
			logger.Infof("table %s doesn't exist", s.RefTable().Name)
			return nil, s.CreateTable()
		}
		table := s.RefTable()
		rows, _ := s.Raw(fmt.Sprintf("SELECT * FROM %s LIMIT 1", table.Name)).QueryRows()
		defer rows.Close()
		columns, _ := rows.Columns()
		addCols := difference(table.FieldNames, columns)
		delCols := difference(columns, table.FieldNames)
		logger.Infof("added cols %v, deleted cols %v", addCols, delCols)
		for _, col := range addCols {
			f := table.GetFiled(col)
			_, err = s.Raw(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table.Name, f.Name, f.Type)).Exec()
			if err != nil {
				return
			}
		}

		if len(delCols) == 0 {
			return
		}

		tmpName := "tmp_" + table.Name
		fieldStr := strings.Join(table.FieldNames, ", ")
		s.Raw(fmt.Sprintf("CREATE TABLE %s AS SELECT %s FROM %s", tmpName, fieldStr, table.Name))
		s.Raw(fmt.Sprintf("DROP TABLE %s", table.Name))
		s.Raw(fmt.Sprintf("ALTER TABLE %s RENAME TO %s", tmpName, table.Name))
		_, err = s.Exec()
		return
	})
	return err
}

// return a - a ∩ b
func difference(a []string, b []string) (diff []string) {
	mb := make(map[string]bool)
	for _, v := range b {
		mb[v] = true
	}

	for _, v := range a {
		if _, ok := mb[v]; !ok {
			diff = append(diff, v)
		}
	}
	return
}
