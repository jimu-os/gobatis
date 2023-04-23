package mapper_example

import (
	"database/sql"
	"fmt"
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

func TestAddOne(t *testing.T) {
	s := model.Student{
		Name:       "test",
		Age:        1,
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	if err := studentMapper.AddOne(s); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestAdds(t *testing.T) {
	var arr []model.Student
	for i := 0; i < 10; i++ {
		s := model.Student{
			Name:       fmt.Sprintf("test_%d", i),
			Age:        i + 2,
			CreateTime: time.Now().Format("2006-01-02 15:04:05"),
		}
		arr = append(arr, s)
	}
	err := studentMapper.Adds(
		map[string]any{
			"arr": arr,
		},
	)
	if err != nil {
		t.Error(err.Error())
		return
	}
}

func TestInsertId(t *testing.T) {
	var count, id int64
	var err error
	s := model.Student{
		Name:       "test",
		Age:        2,
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	if count, id, err = studentMapper.InsertId(s); err != nil {
		t.Error(err.Error())
		return
	}
	t.Log("count:", count, "id:", id)
}

func TestQueryAll(t *testing.T) {
	var stus []model.Student
	var err error
	if stus, err = studentMapper.QueryAll(); err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(stus)
}

func TestQueryPage(t *testing.T) {
	var stus []model.Student
	var count int64
	var err error
	if stus, count, err = studentMapper.QueryPage(); err != nil {
		t.Error(err.Error())
		return
	}
	t.Log("rows:", stus, "count:", count)
}
