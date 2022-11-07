package sgo

import (
	"github.com/beevik/etree"
	"reflect"
	"strings"
)

func StatementElement(element *etree.Element, template string, ctx map[string]any) (string, error) {

	return "", nil
}
func ForElement(element *etree.Element, template string, ctx map[string]any) (string, error) {

	return "", nil
}

func IfElement(element *etree.Element, template string, ctx map[string]any) (string, error) {

	return "", nil
}

func toMap(value any) map[string]any {
	valueOf := reflect.ValueOf(value)
	if valueOf.Kind() == reflect.Pointer {
		valueOf = valueOf.Elem()
		return toMap(valueOf.Interface())
	}
	ctx := make(map[string]any)
	var key string
	var v any
	switch valueOf.Kind() {
	case reflect.Struct:
		for i := 0; i < valueOf.NumField(); i++ {
			field := valueOf.Field(i)
			key = valueOf.Type().Field(i).Name
			key = strings.ToLower(key)
			v = field.Interface()
			if field.Kind() == reflect.Struct || field.Kind() == reflect.Pointer || field.Kind() == reflect.Map {
				v = toMap(v)
			}
			ctx[key] = v
		}
	case reflect.Map:
		mapIter := valueOf.MapRange()
		for mapIter.Next() {
			key = mapIter.Key().Interface().(string)
			vOf := mapIter.Value()
			if vOf.Kind() == reflect.Struct || vOf.Kind() == reflect.Map || vOf.Kind() == reflect.Pointer {
				v = toMap(vOf.Interface())
			}
			ctx[key] = v
		}
	}
	return ctx
}
