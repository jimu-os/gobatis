package gobatis

import (
	"database/sql"
	"gitee.com/aurora-engine/gobatis/obj"
	"reflect"
)

// Null 采用结构体全名模式进行校验
// 通过判断接口接口实现，对后续的自定义结构体解析处理，和空值处理存在冲突，暂时没有处理冲突，接口判断是强制判定，无法让指针和值类型同时验证成功
var Null = map[string]bool{
	TypeKey(sql.NullInt16{}):   true,
	TypeKey(sql.NullInt32{}):   true,
	TypeKey(sql.NullInt64{}):   true,
	TypeKey(sql.NullFloat64{}): true,
	TypeKey(sql.NullBool{}):    true,
	TypeKey(sql.NullString{}):  true,
	TypeKey(sql.NullByte{}):    true,
	TypeKey(sql.NullTime{}):    true,

	TypeKey(&sql.NullInt16{}):   true,
	TypeKey(&sql.NullInt32{}):   true,
	TypeKey(&sql.NullInt64{}):   true,
	TypeKey(&sql.NullFloat64{}): true,
	TypeKey(&sql.NullBool{}):    true,
	TypeKey(&sql.NullString{}):  true,
	TypeKey(&sql.NullByte{}):    true,
	TypeKey(&sql.NullTime{}):    true,

	// 自定义 null 支持
	TypeKey(obj.String{}):  true,
	TypeKey(&obj.String{}): true,
}

// NullConfig 添加 null 数据
// null 数据不会被 databaseToGolang 映射函数处理，相应的如果要在自定义空值中处理特殊的赋值逻辑，通过Scan 接口也可以实现，例如obj.String中处理逻辑
// obj.String 做了对 普通字符串和数据库时间类型的通用匹配
//
//	type String struct {
//		V string
//	}
//
//	func (s *String) Scan(data any) error {
//		if data != nil {
//			switch data.(type) {
//			case time.Time:
//				s.V = data.(time.Time).Format("2006-01-02 15:04:05")
//			case string:
//				s.V = data.(string)
//			case []byte:
//				s.V = string(data.([]byte))
//			}
//		}
//		return nil
//	}
//
// 参考 obj/string.go
func NullConfig(null any) {
	Null[reflect.TypeOf(null).String()] = true
}
