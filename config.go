package gobatis

type DBSqlTemplateFunc func(string) string

var dbTemplate = map[int]DBSqlTemplateFunc{
	MySQL:      defaultSql,
	PostgreSQL: toPgPlaceholder,
	Sqlite:     defaultSql,
}
