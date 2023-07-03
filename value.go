package gobatis

import "strconv"

func elementValue(value any) (v string, flag bool, err error) {
	switch value.(type) {
	case string:
		v = value.(string)
		flag = true
	case int:
		v = strconv.Itoa(value.(int))
	case int64:
		v = strconv.FormatInt(value.(int64), 10)
	case float64:
		v = strconv.FormatFloat(value.(float64), 'f', 2, 64)
	case bool:
		v = strconv.FormatBool(value.(bool))
	default:
		// 其他复杂数据类型
		if handle, err := dataHandle(value); err != nil {
			return "", flag, err
		} else {
			return elementValue(handle)
		}
	}
	return
}
