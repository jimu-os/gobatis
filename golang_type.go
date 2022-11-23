package sgo

import (
	"reflect"
	"time"
)

func init() {
	databaseToGolang = map[string]ToGolang{
		TypeKey(time.Time{}):  TimeData,
		TypeKey(&time.Time{}): TimeDataPointer,
	}
}

// ToGolang 处理数据库从查询结果集中的复杂数据类型的赋值
// value : 是在一个结构体内的字段反射，通过该函数可以对这个字段进行初始化赋值
// data  : 是value对应的具体参数值，可能是字符串，切片，map
type ToGolang func(value reflect.Value, data any) error

// databaseToGolang  定义了 查询结果集对应的 golang 数据映射匹配器
// key : 通过对类型的反射取到的类型名称
// value : 定义了对应该类型的解析逻辑
var databaseToGolang map[string]ToGolang

//golang_type.go 存放 数据库查询结果集映射到 golang中结构体的字段类型处理器

// GolangType 对外提供添加 自定义结果集数据类型解析支持
// key 需要通过 TypeKey 函数获取一个全局唯一的标识符
// dataType 需要提供 对应数据解析逻辑细节可以参考 TimeData 或者 TimeDataPointer
func GolangType(key string, dataType ToGolang) {
	if _, b := databaseToGolang[key]; !b {
		databaseToGolang[key] = dataType
	}
}

// TimeData 时间类型数据
func TimeData(value reflect.Value, data any) error {
	t := data.(string)
	parse, err := time.Parse("2006-04-02 15:04:05", t)
	if err != nil {
		return err
	}
	value.Set(reflect.ValueOf(parse))
	return nil
}

func TimeDataPointer(value reflect.Value, data any) error {
	t := data.(string)
	parse, err := time.Parse("2006-04-02 15:04:05", t)
	if err != nil {
		return err
	}
	value.Set(reflect.ValueOf(&parse))
	return nil
}
