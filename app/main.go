package main

import (
	"fmt"
	"reflect"
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
	//build := &sqlgo.Build{}
	//build.LoadXml("/sql")
	s := new(StuMapper)
	of := reflect.ValueOf(*s)
	fmt.Println(of.String())
	//method := of.NumMethod()
	fmt.Println(of.Kind())
}
