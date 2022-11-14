package sgo

import (
	"fmt"
	"reflect"
	"strings"
)

type Slice struct {
	/*
		实现普通切片的数据处理
	*/
}

func (s Slice) ForEach(value any, template string, separator string) (string, string, []any, error) {
	var v any
	var err error
	var item, itemSql string
	var param []any
	items := make([]string, 0)
	tempSql := make([]string, 0)
	params := make([]any, 0)
	valueOf := reflect.ValueOf(value)
	for i := 0; i < valueOf.Len(); i++ {
		IndexV := valueOf.Index(i)
		v = IndexV.Interface()
		if IndexV.Kind() == reflect.Slice {
			return "", "", nil, fmt.Errorf("'slice' element error")
		}
		if IndexV.Kind() == reflect.Map {
			ctx := v.(map[string]any)
			item, itemSql, param, err = AnalysisForTemplate(template, ctx, nil)
			if err != nil {
				return "", "", nil, err
			}
			items = append(items, item)
			tempSql = append(tempSql, itemSql)
			params = append(params, param...)
			continue
		}
		item, itemSql, param, err = AnalysisForTemplate(template, nil, v)
		if err != nil {
			return "", "", nil, err
		}
		items = append(items, item)
		tempSql = append(tempSql, itemSql)
		params = append(params, param...)
	}
	join := strings.Join(items, separator)
	s2 := strings.Join(tempSql, separator)
	return join, s2, params, nil
}
