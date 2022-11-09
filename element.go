package sgo

import (
	"bytes"
	"fmt"
	"github.com/antonmedv/expr"
	"github.com/beevik/etree"
	"reflect"
	"strings"
)

func StatementElement(element *etree.Element, template string, ctx map[string]any) (string, error) {
	analysisTemplate, err := AnalysisTemplate(template, ctx)
	if err != nil {
		return "", fmt.Errorf("%s,%s,%s", element.Tag, element.SelectAttr("id").Value, err.Error())
	}
	return analysisTemplate, nil
}

func ForElement(element *etree.Element, template string, ctx map[string]any) (string, error) {
	return forElement(element, template, ctx)
}

func IfElement(element *etree.Element, template string, ctx map[string]any) (string, error) {
	return ifElement(element, template, ctx)
}

func forElement(element *etree.Element, template string, ctx map[string]any) (string, error) {
	var slice, item, open, closes, separator, column string
	var attr *etree.Attr
	buf := bytes.Buffer{}

	attr = element.SelectAttr("column")
	if attr == nil {
		return "", fmt.Errorf("%s column is not found", element.Tag)
	}
	column = attr.Value
	if column != "" {
		buf.WriteString(column + " IN ")
	}

	attr = element.SelectAttr("slice")
	if attr == nil {
		return "", fmt.Errorf("%s slice is not found", element.Tag)
	}
	slice = attr.Value
	if slice == "" {
		return "", fmt.Errorf("%s 'slice' Attr not empty", element.Tag)
	}

	attr = element.SelectAttr("item")
	if attr == nil {
		return "", fmt.Errorf("%s item is not found", element.Tag)
	}
	item = attr.Value
	if item == "" {
		return "", fmt.Errorf("%s 'item' Attr not empty", element.Tag)
	}

	attr = element.SelectAttr("open")
	if attr == nil {
		return "", fmt.Errorf("%s open is not found", element.Tag)
	}
	open = attr.Value
	if open == "" {
		open = "("
	}

	attr = element.SelectAttr("close")
	if attr == nil {
		return "", fmt.Errorf("%s close is not found", element.Tag)
	}
	closes = attr.Value
	if closes == "" {
		closes = ")"
	}

	attr = element.SelectAttr("separator")
	if attr == nil {
		return "", fmt.Errorf("%s separator is not found", element.Tag)
	}
	separator = attr.Value
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
	// template 预处理
	unTemplate := UnTemplate(template)
	split := strings.Split(unTemplate[1], ".")
	length := len(split)
	if length > 1 && valueOf.Type().Elem().Kind() != reflect.Map {
		return "", fmt.Errorf("")
	}
	buf.WriteString(open)
	var result string
	// 解析 slice 属性迭代
	combine := Combine{Value: v, Ctx: ctx, Keys: split}
	switch valueOf.Kind() {
	case reflect.Slice, reflect.Array:
		combine.Politic = Slice{}
	case reflect.Struct:
		combine.Politic = Struct{}
	case reflect.Pointer:
		combine.Politic = Pointer{}
	default:
		return "", fmt.Errorf("%s is not list", unTemplate)
	}
	result, err = combine.ForEach()
	if err != nil {
		return "", err
	}
	buf.WriteString(result)
	buf.WriteString(closes)

	return buf.String(), nil
}

func ifElement(element *etree.Element, template string, ctx map[string]any) (string, error) {
	var attr *etree.Attr
	attr = element.SelectAttr("expr")
	if attr == nil {
		return "", fmt.Errorf("%s,attr 'expr' not found", element.Tag)
	}
	exprStr := attr.Value
	if exprStr == "" {
		return "", fmt.Errorf("%s,attr 'expr' value is empty", element.Tag)
	}
	analysisExpr := AnalysisExpr(exprStr)
	compile, err := expr.Compile(analysisExpr)
	if err != nil {
		return "", err
	}
	run, err := expr.Run(compile, ctx)
	if err != nil {
		return "", err
	}
	var flag, f bool
	if flag, f = run.(bool); !f {
		return "", fmt.Errorf("%s,expr result is not bool type", element.Tag)
	}
	if flag {
		analysisTemplate, err := AnalysisTemplate(template, ctx)
		if err != nil {
			return "", fmt.Errorf("%s,template '%s'. %s", element.Tag, template, err.Error())
		}
		return analysisTemplate, nil
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
			if !valueOf.Type().Field(i).IsExported() {
				continue
			}
			key = valueOf.Type().Field(i).Name
			key = strings.ToLower(key)
			v = field.Interface()
			if dataType(key, v, ctx) {
				continue
			}
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
			if dataType(key, v, ctx) {
				continue
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

// 校验复杂数据类型，不是复杂数据类型返回 false 让主程序继续处理，如果是复杂数据类型，应该直接添加到ctx，并返回true
func dataType(key string, value any, ctx map[string]any) bool {

	return false
}

// 模板解析处理复杂数据类型
func dataHandle(value any) string {

	return ""
}

func UnTemplate(template string) []string {
	length := len(template)
	return []string{template[0:1], template[1 : length-1], template[length-1:]}
}

// AnalysisExpr 翻译表达式
func AnalysisExpr(template string) string {
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
					break
				}
			}
			s := template[starIndex+1 : endIndex]
			buf.WriteString(" " + s + " ")
			i = endIndex + 1
			continue
		}
		buf.WriteByte(templateByte[i])
		i++
	}
	return buf.String()
}

// AnalysisTemplate 模板解析器
func AnalysisTemplate(template string, ctx map[string]any) (string, error) {
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
				return "", fmt.Errorf("%s,'%s' not found", template, s)
			}
			switch value.(type) {
			case string:
				buf.WriteString(fmt.Sprintf(" '%s' ", value.(string)))
			case int:
				buf.WriteString(fmt.Sprintf(" %d ", value.(int)))
			case float64:
				buf.WriteString(fmt.Sprintf(" %f ", value.(float64)))
			default:
				// 其他复杂数据类型
				if handle := dataHandle(value); handle != "" {
					buf.WriteString(" " + handle + " ")
				}
			}
			i = endIndex + 1
			continue
		}
		buf.WriteByte(templateByte[i])
		i++
	}
	return buf.String(), nil
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
