package gobatis

import "strconv"

func elementValue(value any) (v string, err error) {
	switch value.(type) {
	case string:
		v = value.(string)
	case int:
		v = strconv.Itoa(value.(int))
	case int64:
		v = strconv.Itoa(int(value.(int64)))
	case float64:
		v = strconv.FormatFloat(value.(float64), 'f', 'g', 64)
	case bool:
		v = strconv.FormatBool(value.(bool))
	default:
		// 其他复杂数据类型
		if handle, e := dataHandle(value); e != nil {
			return "", e
		} else {
			return elementValue(handle)
		}
	}
	return
}
