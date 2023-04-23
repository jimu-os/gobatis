package mapper_example

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

func TestInsert(t *testing.T) {

}
