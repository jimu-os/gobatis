package mapper_example

import (
	"database/sql"
	"gitee.com/aurora-engine/gobatis"
	"gitee.com/aurora-engine/gobatis/examples/mapper_example/model"
	_ "github.com/go-sql-driver/mysql"
	"testing"
	"time"
)

var studentMapper *StudentMapper

func init() {
	studentMapper = &StudentMapper{}
	open, err := sql.Open("mysql", "root:Aurora@2022@tcp(82.157.160.117:3306)/community?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		return
	}
	batis := gobatis.New(open)
	batis.Source("/")
	batis.ScanMappers(studentMapper)
}

func TestInsert(t *testing.T) {
	s := model.Student{
		Id:         0,
		Name:       "test",
		Age:        1,
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}
	err := studentMapper.AddOne(s)
	if err != nil {
		t.Error(err.Error())
		return
	}
}
