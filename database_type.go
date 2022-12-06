package gobatis

import "time"

func init() {
	golangToDatabase = map[string]ToDatabase{
		TypeKey(time.Time{}):  ToDatabaseTime,
		TypeKey(&time.Time{}): ToDatabaseTimePointer,
	}
}

// ToDatabase mapper 中sql解析模板对应的复杂数据据类型解析器
// data : 对应的数据本身
// 对应需要返回一个非结构体的基础数据类型（int float，bool，string） 更具需要构成的实际sql决定，后续的sql解析将自动匹配数据类
type ToDatabase func(data any) (any, error)

var golangToDatabase map[string]ToDatabase

//database_type.go 存放 golang 数据类型对应解析到对应的 数据库字段处理器

// DatabaseType 对外提供添加 自定义sql语句数据类型解析支持
func DatabaseType(key string, dataType ToDatabase) {
	if _, b := golangToDatabase[key]; !b {
		golangToDatabase[key] = dataType
	}
}

func ToDatabaseTime(data any) (any, error) {
	t := data.(time.Time)
	return t.Format("2006-01-02 15:04:05"), nil
}

func ToDatabaseTimePointer(data any) (any, error) {
	t := data.(*time.Time)
	return t.Format("2006-01-02 15:04:05"), nil
}
