package main

import (
	"fmt"
	"orm"

	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	Name string `orm:"PRIMARY KEY"`
	Age  int    `orm:"DEFAULT(18)"`
}

func main() {
	fmt.Println("Hello World!")
	engine, _ := orm.NewEngine("sqlite3", "orm/test.db")
	defer engine.Close()
	s := engine.NewSession()
	affected, _ := s.Insert(&User{"Insert", 18})
	fmt.Println(affected)

}
