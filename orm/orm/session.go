package orm

import (
	"database/sql"
	"errors"
	"fmt"
	"orm/dialect"
	"orm/logger"
	"reflect"
	"strings"
)

type CommonDB interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
}

var _ CommonDB = (*sql.DB)(nil)
var _ CommonDB = (*sql.Tx)(nil)

var (
	ErrModelNotSet    = errors.New("model is not set")
	ErrRecordNotFound = errors.New("record not found")
)

type Session struct {
	db       *sql.DB
	dialect  dialect.Dialect
	tx       *sql.Tx
	refTable *Schema
	sql      strings.Builder
	sqlVars  []any
	clause   Clause
	hook     bool
}

func NewSession(db *sql.DB, dialect dialect.Dialect) *Session {
	return &Session{db: db, dialect: dialect}
}

func (s *Session) DB() CommonDB {
	if s.tx != nil {
		return s.tx
	}
	return s.db
}

func (s *Session) Clear() {
	s.sql.Reset()
	s.sqlVars = nil
	s.clause = Clause{}
}

func (s *Session) Raw(sql string, values ...any) *Session {
	s.sql.WriteString(sql)
	s.sql.WriteString("; ")
	s.sqlVars = append(s.sqlVars, values...)

	return s
}

func (s *Session) Exec() (result sql.Result, err error) {
	defer s.Clear()
	logger.Info(s.sql.String(), s.sqlVars)
	result, err = s.DB().Exec(s.sql.String(), s.sqlVars...)
	if err != nil {
		logger.Error(err)
	}
	return
}

func (s *Session) QueryRow() *sql.Row {
	defer s.Clear()
	query := s.sql.String()
	query = query[:len(query)-2]
	logger.Info(query, s.sqlVars)
	return s.DB().QueryRow(query, s.sqlVars...)
}

func (s *Session) QueryRows() (rows *sql.Rows, err error) {
	defer s.Clear()
	query := s.sql.String()
	query = query[:len(query)-2]
	logger.Info(query, s.sqlVars)
	rows, err = s.DB().Query(query, s.sqlVars...)
	if err != nil {
		logger.Error(err)
	}
	return
}

func (s *Session) Model(value any) *Session {
	if s.refTable == nil || reflect.TypeOf(value) != reflect.TypeOf(s.refTable.Model) {
		s.refTable = Parse(value, s.dialect)
	}
	return s
}

func (s *Session) RefTable() *Schema {
	if s.refTable == nil {
		logger.Error(ErrModelNotSet)
	}
	return s.refTable
}

func (s *Session) CreateTable() error {
	schema := s.refTable
	if schema == nil {
		logger.Error(ErrModelNotSet)
		return ErrModelNotSet
	}

	var columns []string
	for _, field := range schema.Fields {
		columns = append(columns, fmt.Sprintf("%s %s %s", field.Name, field.Type, field.Tag))
	}
	desc := strings.Join(columns, ",")
	_, err := s.Raw(fmt.Sprintf("CREATE TABLE %s (%s)", schema.Name, desc)).Exec()
	return err
}

func (s *Session) DropTable() error {
	schema := s.refTable
	if schema == nil {
		logger.Error(ErrModelNotSet)
		return ErrModelNotSet
	}
	_, err := s.Raw(fmt.Sprintf("DROP TABLE IF EXISTS %s", schema.Name)).Exec()
	return err
}

func (s *Session) HasTable() bool {
	sql, values := s.dialect.TableExitSQL(s.RefTable().Name)
	row := s.Raw(sql, values...).QueryRow()
	var tmp string
	_ = row.Scan(&tmp)
	return tmp == s.RefTable().Name
}

func (s *Session) Insert(values ...any) (int64, error) {
	table := s.Model(values[0]).RefTable()
	s.clause.Set(INSERT, table.Name, table.FieldNames)
	var value any
	if len(values) > 0 {
		value = values[0]
	}
	s.CallMethod(BeforeInsert, value)

	var recordValues []any
	for _, v := range values {
		recordValues = append(recordValues, table.RecordValues(v))
	}
	s.clause.Set(VALUES, recordValues...)

	sql, vars := s.clause.Build(INSERT, VALUES)
	res, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		logger.Error(err)
		return 0, err
	}
	s.CallMethod(AfterInsert, value)
	return res.RowsAffected()
}

func (s *Session) Find(values ...any) error {
	destSlice := reflect.Indirect(reflect.ValueOf(values[0]))
	destType := destSlice.Type().Elem()
	table := s.Model(reflect.New(destType).Elem().Interface()).RefTable()
	s.clause.Set(SELECT, table.Name, table.FieldNames)
	for _, v := range values {
		s.CallMethod(BeforeQuery, v)
	}

	sql, vars := s.clause.Build(SELECT, WHERE, ORDERBY, LIMIT)
	rows, err := s.Raw(sql, vars...).QueryRows()
	if err != nil {
		logger.Error(err)
		return err
	}

	for rows.Next() {
		dest := reflect.New(destType).Elem()
		var vals []any
		for _, name := range table.FieldNames {
			vals = append(vals, dest.FieldByName(name).Addr().Interface())
		}
		if err := rows.Scan(vals...); err != nil {
			logger.Error(err)
			return err
		}
		s.CallMethod(AfterQuery, dest.Addr().Interface())
		destSlice.Set(reflect.Append(destSlice, dest))
	}
	return rows.Close()
}

// map[string]any{"name": "Alice", "age": 16}
// or "Name", "Alice", "Age", 16
func (s *Session) Update(values ...any) (int64, error) {
	var value any
	if len(values) > 0 {
		value = values[0]
	}
	s.CallMethod(BeforeUpdate, value)

	m, ok := values[0].(map[string]interface{})
	if !ok {
		m = make(map[string]any)
		for i := 0; i < len(values); i += 2 {
			m[values[i].(string)] = values[i+1]
		}
	}

	s.clause.Set(UPDATE, s.RefTable().Name, m)
	sql, vars := s.clause.Build(UPDATE, WHERE)
	res, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		logger.Error(err)
		return 0, err
	}
	s.CallMethod(AfterUpdate, value)
	return res.RowsAffected()
}

func (s *Session) Delete(values ...any) (int64, error) {
	s.clause.Set(DELETE, s.RefTable().Name)
	var value any
	if len(values) > 0 {
		value = values[0]
	}
	s.CallMethod(BeforeDelete, value)

	sql, vars := s.clause.Build(DELETE, WHERE)
	res, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		logger.Error(err)
		return 0, err
	}
	s.CallMethod(AfterDelete, value)
	return res.RowsAffected()
}

func (s *Session) Count() (int64, error) {
	s.clause.Set(COUNT, s.RefTable().Name)
	sql, vars := s.clause.Build(COUNT, WHERE)
	row := s.Raw(sql, vars...).QueryRow()
	var count int64
	if err := row.Scan(&count); err != nil {
		logger.Error(err)
		return 0, err
	}
	return count, nil
}

func (s *Session) First(value any) error {
	dest := reflect.Indirect(reflect.ValueOf(value))
	destSlice := reflect.New(reflect.SliceOf(dest.Type())).Elem()
	err := s.Limit(1).Find(destSlice.Addr().Interface())
	if err != nil {
		return err
	}
	if destSlice.Len() == 0 {
		return ErrRecordNotFound
	}
	dest.Set(destSlice.Index(0))
	return nil
}

func (s *Session) Last(value any) error {
	dest := reflect.Indirect(reflect.ValueOf(value))
	destSlice := reflect.New(reflect.SliceOf(dest.Type())).Elem()
	err := s.OrderBy("id DESC").Limit(1).Find(destSlice.Addr().Interface())
	if err != nil {
		return err
	}
	if destSlice.Len() == 0 {
		return ErrRecordNotFound
	}
	dest.Set(destSlice.Index(0))
	return nil
}

func (s *Session) Limit(num int) *Session {
	s.clause.Set(LIMIT, num)
	return s
}

func (s *Session) OrderBy(order string) *Session {
	s.clause.Set(ORDERBY, order)
	return s
}

func (s *Session) Where(query string, args ...any) *Session {
	var vars []any
	s.clause.Set(WHERE, append(append(vars, query), args...)...)
	return s
}

func (s *Session) CallMethod(method int, value any) {
	if !s.hook {
		return
	}

	var dest reflect.Value
	if value == nil {
		dest = reflect.ValueOf(s.RefTable().Model)
	} else {
		dest = reflect.ValueOf(value)
	}

	var err error
	switch method {
	case BeforeQuery:
		if i, ok := dest.Interface().(IBeforeQuery); ok {
			err = i.BeforeQuery(s)
		}
	case AfterQuery:
		if i, ok := dest.Interface().(IAfterQuery); ok {
			err = i.AfterQuery(s)
		}
	case BeforeInsert:
		if i, ok := dest.Interface().(IBeforeInsert); ok {
			err = i.BeforeInsert(s)
		}
	case AfterInsert:
		if i, ok := dest.Interface().(IAfterInsert); ok {
			err = i.AfterInsert(s)
		}
	case BeforeUpdate:
		if i, ok := dest.Interface().(IBeforeUpdate); ok {
			err = i.BeforeUpdate(s)
		}
	case AfterUpdate:
		if i, ok := dest.Interface().(IAfterUpdate); ok {
			err = i.AfterUpdate(s)
		}
	case BeforeDelete:
		if i, ok := dest.Interface().(IBeforeDelete); ok {
			err = i.BeforeDelete(s)
		}
	case AfterDelete:
		if i, ok := dest.Interface().(IAfterDelete); ok {
			err = i.AfterDelete(s)
		}
	default:
		err = errors.New("invalid method")
	}

	if err != nil {
		logger.Error(err)
	}
}

func (s *Session) EnableHook() *Session {
	s.hook = true
	return s
}

func (s *Session) Begin() (err error) {
	logger.Info("begin transaction")
	if s.tx, err = s.db.Begin(); err != nil {
		logger.Error(err)
	}
	return
}

func (s *Session) Commit() (err error) {
	logger.Info("commit transaction")
	if err = s.tx.Commit(); err != nil {
		logger.Error(err)
	}
	return
}

func (s *Session) Rollback() (err error) {
	logger.Info("rollback transaction")
	if err = s.tx.Rollback(); err != nil {
		logger.Error(err)
	}
	return
}
