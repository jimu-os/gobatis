package sgo

import (
	"bytes"
	"fmt"
	"github.com/beevik/etree"
	"reflect"
	"strconv"
	"strings"
)

func StatementElement(element *etree.Element, template string, ctx map[string]any) (string, error) {
	buf := bytes.Buffer{}
	template = strings.TrimSpace(template)
	templateByte := []byte(template)
	starIndex := 0
	for i := starIndex; i < len(templateByte); {
		if templateByte[i] == '{' {
			starIndex = i
			endIndex := i
			for j := starIndex; j < len(templateByte); j++ {
				if templateByte[j] == '}' {
					endIndex = j
				}
			}
			s := template[starIndex+1 : endIndex]
			split := strings.Split(s, ".")
			value, err := ctxValue(ctx, split)
			if err != nil {
				return "", err
			}
			switch value.(type) {
			case string:
				buf.WriteString(" '" + value.(string) + "' ")
			case int:
				buf.WriteString(fmt.Sprintf(" %d ", value.(int)))
			case float64:
				buf.WriteString(fmt.Sprintf(" %f ", value.(float64)))
			}
			i = endIndex + 1
			continue
		}
		buf.WriteByte(templateByte[i])
		i++
	}
	return buf.String(), nil
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
	t := UnTemplate(slice)
	keys := strings.Split(t[1], ".")
	v, err := ctxValue(ctx, keys)
	if err != nil {
		return "", err
	}
	valueOf := reflect.ValueOf(v)
	if valueOf.Kind() != reflect.Slice {
		return "", fmt.Errorf("%s 'slice' value is not slice", element.Tag)
	}
	// template 预处理
	unTemplate := UnTemplate(template)
	split := strings.Split(unTemplate[1], ".")
	length := len(split)
	if length > 1 && valueOf.Type().Elem().Kind() != reflect.Map {
		return "", fmt.Errorf("")
	}
	buf.WriteString(open)
	// 解析 slice 属性迭代
	items := []string{}
	for i := 0; i < valueOf.Len(); i++ {
		IndexV := valueOf.Index(i)
		if length == 1 {
			if IndexV.Kind() == reflect.Slice || IndexV.Kind() == reflect.Map {
				return "", fmt.Errorf("'slice' element error")
			}
			a := IndexV.Interface()
			switch a.(type) {
			case int:
				vi := a.(int)
				itoa := strconv.Itoa(vi)
				items = append(items, itoa)
			case string:
				items = append(items, a.(string))
			case float64:
				float := strconv.FormatFloat(a.(float64), 'f', 'x', 64)
				items = append(items, float)
			default:
				return "", fmt.Errorf("")
			}
		} else {
			if IndexV.Kind() != reflect.Map {
				return "", fmt.Errorf("'slice' element error")
			}
			tctx := IndexV.Interface().(map[string]any)
			cv, err := ctxValue(tctx, split[1:])
			if err != nil {
				return "", err
			}
			switch cv.(type) {
			case int:
				vi := cv.(int)
				itoa := strconv.Itoa(vi)
				items = append(items, itoa)
			case string:
				items = append(items, cv.(string))
			case float64:
				float := strconv.FormatFloat(cv.(float64), 'f', 'x', 64)
				items = append(items, float)
			default:
				return "", fmt.Errorf("")
			}
		}
	}
	join := strings.Join(items, separator)
	buf.WriteString(join)
	buf.WriteString(closes)

	return buf.String(), nil
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
			if !valueOf.Type().Field(i).IsExported() {
				continue
			}
			key = valueOf.Type().Field(i).Name
			key = strings.ToLower(key)
			v = field.Interface()
			if field.Kind() == reflect.Slice {
				v = filedToMap(v)
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
			v = vOf.Interface()
			if vOf.Kind() == reflect.Interface {
				if vOf.Elem().Kind() == reflect.Slice {
					if vOf.Elem().Type().Elem().Kind() == reflect.Struct || vOf.Elem().Type().Elem().Kind() == reflect.Pointer || vOf.Elem().Type().Elem().Kind() == reflect.Map {
						v = filedToMap(v)
					}
				}
				if vOf.Elem().Kind() == reflect.Struct || vOf.Elem().Kind() == reflect.Map || vOf.Elem().Kind() == reflect.Pointer {
					v = toMap(v)
				}
			}
			if vOf.Kind() == reflect.Slice {
				v = filedToMap(v)
			}
			if vOf.Kind() == reflect.Struct || vOf.Kind() == reflect.Map || vOf.Kind() == reflect.Pointer {
				v = toMap(v)
			}
			ctx[key] = v
		}
	}
	return ctx
}

func filedToMap(value any) []map[string]any {
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
					vals = filedToMap(v.Interface())
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

func UnTemplate(template string) []string {
	length := len(template)
	return []string{template[0:1], template[1 : length-1], template[length-1:]}
}

// 上下文中取数据
func ctxValue(ctx map[string]any, keys []string) (any, error) {
	kl := len(keys)
	var v any
	b := false
	for i := 0; i < kl; i++ {
		k := keys[i]
		if i == kl-1 {
			if v, b = ctx[k]; !b {
				return nil, fmt.Errorf("'slice' key %s not find ", k)
			}
		} else {
			if v, b = ctx[k]; !b {
				return nil, fmt.Errorf("'slice' key %s not find ", k)
			}
			if ctx, b = v.(map[string]any); !b {
				return nil, fmt.Errorf("'%s' is not map or struct", k)
			}
		}
	}
	return v, nil
}
