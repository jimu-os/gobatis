package main

import (
	"fmt"
	"gitee.com/aurora-engine/sgo"
)

func main() {
	ctx := map[string]any{
		"arr":  []int{1, 2, 3, 4},
		"name": "aaa",
	}
	sgo := sgo.NewSgo()
	sgo.LoadXml("/")
	sql, err := sgo.Sql("user.find", ctx)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(sql)
}
