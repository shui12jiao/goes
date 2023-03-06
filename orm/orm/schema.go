package orm

import (
	"go/ast"
	"orm/dialect"
	"reflect"
)

type Filed struct {
	Name string
	Type string
	Tag  string
}

type Schema struct {
	Model      any
	Name       string
	Fields     []*Filed
	FieldNames []string
	fieldMap   map[string]*Filed
}

func (s *Schema) GetFiled(name string) *Filed {
	return s.fieldMap[name]
}

func (s *Schema) RecordValues(dest any) (fieldValues []any) {
	destValue := reflect.Indirect(reflect.ValueOf(dest))
	for _, field := range s.Fields {
		fieldValues = append(fieldValues, destValue.FieldByName(field.Name).Interface())
	}
	return
}

func Parse(dest any, d dialect.Dialect) *Schema {
	typ := reflect.Indirect(reflect.ValueOf(dest)).Type()
	schema := &Schema{
		Model:    dest,
		Name:     typ.Name(),
		fieldMap: make(map[string]*Filed),
	}

	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if !f.Anonymous && ast.IsExported(f.Name) {

			filed := &Filed{
				Name: f.Name,
				Type: d.TypeMap(f.Type),
				Tag:  f.Tag.Get("orm"),
			}
			schema.Fields = append(schema.Fields, filed)
			schema.FieldNames = append(schema.FieldNames, f.Name)
			schema.fieldMap[f.Name] = filed
		}
	}
	return schema
}
