package gobatis

import (
	"database/sql"
	"gitee.com/aurora-engine/gobatis/obj"
	"reflect"
)

var Null = map[string]bool{
	reflect.TypeOf(sql.NullInt16{}).String():    true,
	reflect.TypeOf(sql.NullInt32{}).String():    true,
	reflect.TypeOf(sql.NullInt64{}).String():    true,
	reflect.TypeOf(sql.NullFloat64{}).String():  true,
	reflect.TypeOf(sql.NullBool{}).String():     true,
	reflect.TypeOf(sql.NullString{}).String():   true,
	reflect.TypeOf(sql.NullByte{}).String():     true,
	reflect.TypeOf(sql.NullTime{}).String():     true,
	reflect.TypeOf(&sql.NullInt16{}).String():   true,
	reflect.TypeOf(&sql.NullInt32{}).String():   true,
	reflect.TypeOf(&sql.NullInt64{}).String():   true,
	reflect.TypeOf(&sql.NullFloat64{}).String(): true,
	reflect.TypeOf(&sql.NullBool{}).String():    true,
	reflect.TypeOf(&sql.NullString{}).String():  true,
	reflect.TypeOf(&sql.NullByte{}).String():    true,
	reflect.TypeOf(&sql.NullTime{}).String():    true,

	// 自定义 null 支持
	reflect.TypeOf(obj.String{}).String():  true,
	reflect.TypeOf(&obj.String{}).String(): true,
}

// NullConfig 添加 null
// Null 数据不会被 databaseToGolang 映射函数处理
func NullConfig(null any) {
	Null[reflect.TypeOf(null).String()] = true
	databaseToGolang[TypeKey(null)] = nil
}
