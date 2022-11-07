package sgo

import (
	"bytes"
	"fmt"
	"github.com/beevik/etree"
	"reflect"
	"strings"
)

func StatementElement(element *etree.Element, template string, ctx map[string]any) (string, error) {

	return "", nil
}

func ForElement(element *etree.Element, template string, ctx map[string]any) (string, error) {
	switch element.Tag {
	case "for":
		return forElement(element, template, ctx)
	case "select", "update", "delete", "insert":
		return StatementElement(element, template, ctx)
	}
	return "", nil
}

func IfElement(element *etree.Element, template string, ctx map[string]any) (string, error) {
	switch element.Tag {
	case "if":
	case "select", "update", "delete", "insert":
		return StatementElement(element, template, ctx)
	}
	return "", nil
}

func forElement(element *etree.Element, template string, ctx map[string]any) (string, error) {
	var slice, item, open, closes, separator, key string
	buf := bytes.Buffer{}
	key = element.SelectAttr("key").Value
	if key != "" {
		buf.WriteString(key + " IN ")
	}
	slice = element.SelectAttr("slice").Value
	if slice == "" {
		return "", fmt.Errorf("%s 'slice' Attr not empty", element.Tag)
	}
	item = element.SelectAttr("item").Value
	if item == "" {
		return "", fmt.Errorf("%s 'item' Attr not empty", element.Tag)
	}
	open = element.SelectAttr("open").Value
	if open == "" {
		open = "("
	}
	closes = element.SelectAttr("close").Value
	if closes == "" {
		closes = ")"
	}
	separator = element.SelectAttr("separator").Value
	if separator == "" {
		separator = ","
	}

	// 上下文中取出 数据
	keys := strings.Split(slice, ".")
	kl := len(keys)
	var v any
	b := false
	for i := 0; i < kl; i++ {
		k := keys[i]
		if i == kl-1 {
			if v, b = ctx[k]; !b {
				return "", fmt.Errorf("%s 'slice' key %s not find ", element.Tag, k)
			}
		} else {
			if v, b = ctx[k]; !b {
				return "", fmt.Errorf("%s 'slice' key %s not find ", element.Tag, k)
			}
			ctx = v.(map[string]any)
		}
	}
	valueOf := reflect.ValueOf(v)
	if valueOf.Kind() != reflect.Slice {
		return "", fmt.Errorf("%s 'slice' value is not slice", element.Tag)
	}

	return "", nil
}

// 把 map 或者 结构体完全转化为 map[any]
func toMap(value any) map[string]any {
	valueOf := reflect.ValueOf(value)
	if valueOf.Kind() != reflect.Map && valueOf.Kind() != reflect.Struct && valueOf.Kind() != reflect.Pointer {
		return map[string]any{}
	}
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
			if field.Kind() == reflect.Slice {
				v = structToMap(v)
			}
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
			if vOf.Kind() == reflect.Slice {
				v = structToMap(vOf.Interface())
			}
			if vOf.Kind() == reflect.Struct || vOf.Kind() == reflect.Map || vOf.Kind() == reflect.Pointer {
				v = toMap(vOf.Interface())
			}
			ctx[key] = v
		}
	}
	return ctx
}

func structToMap(value any) []map[string]any {
	valueOf := reflect.ValueOf(value)
	elem := valueOf.Type().Elem()
	arr := make([]map[string]any, 0)
	length := valueOf.Len()
	switch elem.Kind() {
	case reflect.Struct, reflect.Pointer:
		for i := 0; i < length; i++ {
			val := valueOf.Index(i)
			m := toMap(val.Interface())
			arr = append(arr, m)
		}
	case reflect.Map:
		for i := 0; i < length; i++ {
			val := valueOf.Index(i)
			iter := val.MapRange()
			m := map[string]any{}
			for iter.Next() {
				key := iter.Key().Interface().(string)
				v := iter.Value()
				var vals any
				if v.Kind() == reflect.Slice {
					vals = structToMap(v.Interface())
				}
				if v.Kind() == reflect.Struct || v.Kind() == reflect.Pointer || v.Kind() == reflect.Map {
					vals = toMap(v.Interface())
				}
				m[key] = vals
				arr = append(arr, m)
			}
		}
	}
	return arr
}
