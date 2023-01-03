package gobatis

import (
	"database/sql"
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
}
