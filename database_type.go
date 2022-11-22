package sgo

import (
	"reflect"
)

func init() {
	golangToDatabase = map[string]ToDatabase{}
}

// ToDatabase mapper 中sql解析模板对应的复杂数据据类型解析器
// data : 对应的数据本身
// 对应需要返回一个非结构体的基础数据类型（int float，bool，string）
type ToDatabase func(data any) any

var golangToDatabase map[string]ToDatabase

// TypeKey 通过反射得到一个类型的类型字符串, 适用于普通类型
func TypeKey(t any) string {
	return BaseTypeKey(reflect.ValueOf(t))
}

//database_type.go 存放 golang 数据类型对应解析到对应的 数据库字段处理器

// DatabaseType 对外提供添加 自定义sql语句数据类型解析支持
func DatabaseType(key string, dataType ToDatabase) {
	if _, b := golangToDatabase[key]; !b {
		golangToDatabase[key] = dataType
	}
}
