package gobatis

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/antonmedv/expr"
	"github.com/beevik/etree"
	"reflect"
	"strconv"
	"strings"
)

func StatementElement(element *etree.Element, template string, ctx map[string]any) (string, string, []any, error) {
	analysisTemplate, t, param, err := AnalysisTemplate(template, ctx)
	if err != nil {
		return "", "", param, fmt.Errorf("%s,%s,%s", element.Tag, element.SelectAttr("id").Value, err.Error())
	}
	return analysisTemplate, t, param, nil
}

func ForElement(element *etree.Element, template string, ctx map[string]any) (string, string, []any, error) {
	return forElement(element, template, ctx)
}

func IfElement(element *etree.Element, template string, ctx map[string]any) (string, string, []any, error) {
	return ifElement(element, template, ctx)
}

func forElement(element *etree.Element, template string, ctx map[string]any) (string, string, []any, error) {
	var slice, open, closes, column, keys string
	var attr *etree.Attr
	buf := bytes.Buffer{}
	separator := ","
	templateBuf := bytes.Buffer{}
	params := make([]any, 0)
	if attr = element.SelectAttr("column"); attr != nil {
		column = attr.Value
	}
	if attr = element.SelectAttr("slice"); attr != nil {
		slice = attr.Value
	}
	if attr = element.SelectAttr("open"); attr != nil {
		open = attr.Value
	}
	if attr = element.SelectAttr("close"); attr != nil {
		closes = attr.Value
	}
	if attr = element.SelectAttr("separator"); attr != nil {
		separator = attr.Value
	}
	if column != "" {
		buf.WriteString(column + " IN ")
		templateBuf.WriteString(column + " IN ")
	}
	// 上下文中取出 数据

	keys = UnTemplate(slice)
	key := strings.Split(keys, ".")
	// 上下文参数中找到 keys 的值 v 可能是 切片 数组，也可能是自定义的 List 数据类型等
	v, err := ctxValue(ctx, key)
	if err != nil {
		return "", "", nil, err
	}
	valueOf := reflect.ValueOf(v)
	if open != "" {
		buf.WriteString(open)
		templateBuf.WriteString(open)
	}
	var result, temp string
	var param []any
	// 解析 slice 属性迭代
	combine := Combine{Value: v, Template: template, Separator: separator}
	switch valueOf.Kind() {
	case reflect.Slice, reflect.Array:
		combine.Politic = Slice{}
	default:
		combine.Politic = Other{}
	}
	result, temp, param, err = combine.ForEach()
	if err != nil {
		return "", "", nil, err
	}
	params = append(params, param...)
	buf.WriteString(result)
	templateBuf.WriteString(temp)
	if closes != "" {
		buf.WriteString(closes)
		templateBuf.WriteString(closes)
	}
	return buf.String(), templateBuf.String(), params, nil
}

func ifElement(element *etree.Element, template string, ctx map[string]any) (string, string, []any, error) {
	var attr *etree.Attr
	attr = element.SelectAttr("expr")
	if attr == nil {
		return "", "", nil, fmt.Errorf("%s,attr 'expr' not found", element.Tag)
	}
	exprStr := attr.Value
	if exprStr == "" {
		return "", "", nil, fmt.Errorf("%s,attr 'expr' value is empty", element.Tag)
	}
	analysisExpr := AnalysisExpr(exprStr)
	compile, err := expr.Compile(analysisExpr)
	if err != nil {
		return "", "", nil, err
	}
	run, err := expr.Run(compile, ctx)
	if err != nil {
		return "", "", nil, err
	}
	var flag, f bool
	if flag, f = run.(bool); !f {
		return "", "", nil, fmt.Errorf("%s,expr result is not bool type", element.Tag)
	}
	if flag {
		analysisTemplate, t, param, err := AnalysisTemplate(template, ctx)
		if err != nil {
			return "", t, param, fmt.Errorf("%s,template '%s'. %s", element.Tag, template, err.Error())
		}
		return analysisTemplate, t, param, nil
	}
	return "", "", nil, nil
}

// 把 map 或者 结构体完全转化为 map[any]
func toMap(value any) map[string]any {
	if value == nil {
		return nil
	}
	valueOf := reflect.ValueOf(value)
	if valueOf.Kind() != reflect.Map && valueOf.Kind() != reflect.Struct && valueOf.Kind() != reflect.Pointer {
		return map[string]any{}
	}
	if valueOf.Kind() == reflect.Pointer {
		valueOf = valueOf.Elem()
		return toMap(valueOf.Interface())
	}
	ctx := make(map[string]any)
	switch valueOf.Kind() {
	case reflect.Struct:
		structToMap(valueOf, ctx)
	case reflect.Map:
		mapToMap(valueOf, ctx)
	}
	return ctx
}

func structToMap(value reflect.Value, ctx map[string]any) {
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		if !value.Type().Field(i).IsExported() {
			continue
		}
		FiledType := value.Type().Field(i)
		key := FiledType.Name
		if tag, b := FiledType.Tag.Lookup("name"); b && tag != "" {
			key = tag
		}
		v := field.Interface()
		if dataType(v) {
			ctx[key] = v
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
}

func mapToMap(value reflect.Value, ctx map[string]any) {
	mapIter := value.MapRange()
	for mapIter.Next() {
		key := mapIter.Key().Interface().(string)
		vOf := mapIter.Value()
		v := vOf.Interface()
		if dataType(v) {
			ctx[key] = v
			continue
		}
		if vOf.Kind() == reflect.Interface {
			if vOf.Elem().Kind() == reflect.Slice {
				if vOf.Elem().Type().Elem().Kind() == reflect.Struct ||
					vOf.Elem().Type().Elem().Kind() == reflect.Pointer ||
					vOf.Elem().Type().Elem().Kind() == reflect.Map {
					v = filedToMap(v)
				}
				// todo 修复切片中存储的 interface{} 类型的具体指向,基础数据类型不做任何处理
				if vOf.Elem().Type().Elem().Kind() == reflect.Interface {
					if vOf.Elem().Len() > 0 {
						index := vOf.Elem().Index(0)
						switch index.Elem().Type().Kind() {
						case reflect.Struct, reflect.Pointer, reflect.Map:
							v = filedToMap(v)
						}
					}
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

func filedToMap(value any) []map[string]any {
	valueOf := reflect.ValueOf(value)
	elem := valueOf.Type().Elem()
	arr := make([]map[string]any, 0)
	length := valueOf.Len()
	switch elem.Kind() {
	case reflect.Struct, reflect.Pointer, reflect.Interface:
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
				vals = v.Interface()
				if v.Kind() == reflect.Slice {
					vals = filedToMap(v.Interface())
				}
				if v.Kind() == reflect.Struct || v.Kind() == reflect.Pointer || v.Kind() == reflect.Map {
					vals = toMap(v.Interface())
				}
				m[key] = vals
			}
			arr = append(arr, m)
		}
	}
	return arr
}

// dataType 校验 map 转化，注册了 DatabaseType 的数据 将跳过数据的转化，保留原始类型
// 校验复杂数据类型，不是复杂数据类型返回 false 让主程序继续处理，如果是复杂数据类型，应该直接添加到ctx，并返回true
func dataType(value any) bool {
	typeKey := TypeKey(value)
	if _, b := golangToDatabase[typeKey]; b {
		return b
	}
	return false
}

// 模板解析处理复杂数据类型
func dataHandle(value any) (any, error) {
	// TODO 处理复杂数据类型解析，更具数据解析器得到的数据
	key := TypeKey(value)
	database := golangToDatabase[key]
	if database == nil {
		return nil, errors.New("unknown data processing")
	}
	result, err := database(value)
	if err != nil {
		return "", err
	}
	return result, nil
}

// UnTemplate 解析 {xx} 模板 解析为三个部分 ["{","xx","}"]
func UnTemplate(template string) string {
	if length := len(template); length > 3 && (template[0:1] == "{" && template[length-1:] == "}") {
		return template[1 : length-1]
	}
	panic("Failed to resolve template format errors. Procedure")
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
func AnalysisTemplate(template string, ctx map[string]any) (string, string, []any, error) {
	params := []any{}
	buf := bytes.Buffer{}
	templateBuf := bytes.Buffer{}
	template = strings.TrimSpace(template)
	templateByte := []byte(template)
	starIndex := 0
	for i := starIndex; i < len(templateByte); {
		if chat, types, flag := checkTemplateArgs(templateByte[i]); flag {
			// 解析到模版参数
			starIndex = i
			endIndex := i
			for j := starIndex; j < len(templateByte); j++ {
				if templateByte[j] == chat {
					endIndex = j
					break
				}
			}
			// 上面如果没有解析到模版参数结尾标记，则直表示模版存在错误需要报错
			if starIndex == endIndex {
				panic(fmt.Sprintf("%s Template format error\n", template[:starIndex+1]))
			}
			s := template[starIndex+1 : endIndex]
			split := strings.Split(s, ".")
			value, err := ctxValue(ctx, split)
			if err != nil {
				return "", "", params, fmt.Errorf("%s,'%s' not found", template, s)
			}
			//封装 数据解析
			v, flag, err := elementValue(value)
			if err != nil {
				return "", "", nil, err
			}
			if flag && types {
				buf.WriteString("'" + v + "'")
			} else {
				buf.WriteString(v)
			}
			if templateByte[i] == '[' && chat == ']' {
				// 对 [] 参数做特殊处理 不让 在需要执行的 sql 模版中语句中出现
				templateBuf.WriteString(v)
				i = endIndex + 1
				continue
			}
			templateBuf.WriteString("?")
			params = append(params, value)
			i = endIndex + 1
			continue
		}
		buf.WriteByte(templateByte[i])
		templateBuf.WriteByte(templateByte[i])
		i++
	}
	return buf.String(), templateBuf.String(), params, nil
}

// 上下文中取数据
func ctxValue(ctx map[string]any, keys []string) (any, error) {
	if ctx == nil {
		return nil, errors.New("ctx is nil")
	}
	if keys == nil {
		return nil, nil
	}
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

// for迭代 上下文中取数据
func sliceCtxValue(ctx map[string]any, keys []string) (any, error) {
	if ctx == nil {
		return nil, errors.New("ctx is nil")
	}
	var v any
	b := false
	keys = keys[1:]
	kl := len(keys)
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

// 合并 map 吧 src 下的内容合并到 target 下，同名的 属性将被覆盖
func mergeMap(target, src map[string]any) {
	for k, v := range src {
		target[k] = v
	}
}

// 检查解析的字符 是否是模版参数
// 返回值1 返回模版参数类型
// 返回值2 表示此字符是否是模版参数的开始字符
// 当前模版参数 仅有 {} [] 两种
func checkTemplateArgs(b byte) (byte, bool, bool) {
	if b == '{' {
		return '}', true, true
	}

	if b == '[' {
		return ']', false, true
	}
	return b, false, false
}

// 转换为 postgres 占位符
func toPgPlaceholder(sqlStr string) string {
	builder := strings.Builder{}
	index := 1
	for _, v := range sqlStr {
		if v == '?' {
			builder.WriteString("$")
			builder.WriteString(strconv.Itoa(index))
			index++
		} else {
			builder.WriteString(string(v))
		}
	}
	return builder.String()
}
