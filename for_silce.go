package sgo

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type Slice struct {
	/*
		实现普通切片的数据处理
	*/
}

func (s Slice) ForEach(value any, ctx map[string]any, keys []string) (string, error) {
	items := make([]string, 0)
	length := len(keys)
	valueOf := reflect.ValueOf(value)
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
			cv, err := ctxValue(tctx, keys[1:])
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
			}
		}
	}
	join := strings.Join(items, ",")
	return join, nil
}
