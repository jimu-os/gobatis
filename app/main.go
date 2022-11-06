package main

import (
	"fmt"
	"gitee.com/aurora-engine/sqlgo"
)

type Student struct {
	Id   int
	Name string
	Age  int
}
type StuMapper interface {
	SelectStudent() Student
}

type Mapper func(any)

func main() {
	build := sqlgo.NewBuild()
	build.LoadXml("/sql")
	sql, err := build.Sql("user.SelectStudent", nil)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(sql)
}
