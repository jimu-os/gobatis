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

func (s Slice) ForEach(value any, template string, separator string) (string, error) {
	var v any
	var err error
	var item string
	items := make([]string, 0)
	valueOf := reflect.ValueOf(value)
	for i := 0; i < valueOf.Len(); i++ {
		IndexV := valueOf.Index(i)
		v = IndexV.Interface()
		if IndexV.Kind() == reflect.Slice {
			return "", fmt.Errorf("'slice' element error")
		}
		if IndexV.Kind() == reflect.Map {
			ctx := v.(map[string]any)
			item, err = AnalysisForTemplate(template, ctx, nil)
			if err != nil {
				return "", err
			}
			items = append(items, item)
			continue
		}
		item, err = AnalysisForTemplate(template, nil, v)
		if err != nil {
			return "", err
		}
		items = append(items, item)
	}
	join := strings.Join(items, separator)
	return join, nil
}
