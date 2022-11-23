package sgo

import (
	"fmt"
	"reflect"
	"strings"
)

// TypeKey 通过反射得到一个类型的类型字符串, 适用于普通类型
func TypeKey(t any) string {
	return BaseTypeKey(reflect.ValueOf(t))
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
