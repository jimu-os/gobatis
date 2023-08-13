package gobatis

import "strconv"

func elementValue(value any) (v string, flag bool, err error) {
	switch vt := value.(type) {
	case string:
		v = vt
		// 字符串 此处就返回 true 表示 所有字符串都会被添加上 '' 单引号
		flag = true
	case int:
		v = strconv.Itoa(vt)
	case int64:
		v = strconv.FormatInt(vt, 10)
	case float64:
		v = strconv.FormatFloat(vt, 'f', 2, 64)
	case bool:
		v = strconv.FormatBool(vt)
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
