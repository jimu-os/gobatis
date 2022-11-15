package sgo

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// 加载 基础类型解析支持
// 泛型的基础类型解析需要手动导入 key 格式为:  包全名-类型字符串
func init() {
	BaseType = map[string]DataType{
		TypeKey(time.Time{}):  TimeData,
		TypeKey(&time.Time{}): TimeDataPointer,
	}
}

// DataType 函数定义反射赋值逻辑
// value : 是在一个结构体内的字段反射，通过该函数可以对这个字段进行初始化赋值
// data  : 是value对应的具体参数值，可能是字符串，切片，map
type DataType func(value reflect.Value, data any) error

// BaseType 存储了请求参数解析过程中对结构体内部字段类型的自定义支持，添加到 Type 中的类型在 控制器参数校验时候会自动跳过
// key : 通过对类型的反射取到的类型名称
// value : 定义了对应该类型的解析逻辑
var BaseType map[string]DataType

// ValueType 对外提供添加 自定义数据类型解析支持
// key 需要通过 TypeKey 函数获取一个全局唯一的标识符
// dataType 需要提供 对应数据解析逻辑细节可以参考 AuroraQueuePointerType 或者 AuroraStackPointerType
func ValueType(key string, dataType DataType) {
	if _, b := BaseType[key]; !b {
		BaseType[key] = dataType
	}
}

// TypeKey 通过反射得到一个类型的类型字符串, 适用于普通类型
func TypeKey(t any) string {
	typeOf := reflect.TypeOf(t)
	baseType := ""
	if typeOf.Kind() == reflect.Ptr {
		baseType = fmt.Sprintf("%s-%s", typeOf.Elem().PkgPath(), typeOf.String())
	} else {
		baseType = fmt.Sprintf("%s-%s", typeOf.PkgPath(), typeOf.String())
	}
	return baseType
}

// BaseTypeKey 通过 BaseTypeKey 得到的变量默认全包名对泛型参数进行特殊处理的，不会加上类型中的 [xxx]定义部分信息
func BaseTypeKey(v reflect.Value) string {
	baseType := ""
	if v.Kind() == reflect.Ptr {
		baseType = fmt.Sprintf("%s-%s", v.Type().Elem().PkgPath(), v.Type().String())
	} else {
		baseType = fmt.Sprintf("%s-%s", v.Type().PkgPath(), v.Type().String())
	}
	if index := strings.Index(baseType, "["); index != -1 {
		baseType = baseType[:index]
	}
	return baseType
}

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
