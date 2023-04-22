package count_example

import (
	"database/sql"
	"gitee.com/aurora-engine/gobatis"
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

var countMapper *StudentMapper

func init() {
	countMapper = &StudentMapper{}
	open, err := sql.Open("mysql", "root:Aurora@2022@tcp(82.157.160.117:3306)/community?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		return
	}
	batis := gobatis.New(open)
	batis.Source("/")
	batis.ScanMappers(countMapper)
}

func TestCount(t *testing.T) {
	students, count, err := countMapper.Select()
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(students)
	t.Log(count)
}
