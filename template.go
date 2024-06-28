package gobatis

import (
	"strconv"
	"strings"
)

func defaultSql(sqlStr string) string {
	return sqlStr
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
