package gobatis

import "strings"

const (
	gt = "&gt;"
	lt = "&lt;"
)

func conditional(sql string) string {
	sql = strings.Replace(sql, gt, ">", -1)
	sql = strings.Replace(sql, lt, "<", -1)
	return sql
}
