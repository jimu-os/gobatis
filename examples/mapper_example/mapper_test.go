package mapper_example

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jimu-os/gobatis"
	"github.com/jimu-os/gobatis/examples/mapper_example/model"
	"strconv"
	"strings"
	"testing"
	"time"
)

var err error
var open *sql.DB
var studentMapper = &StudentMapper{}
var tag = &TagTestMapper{}

func init() {
	open, err = sql.Open("mysql", "root:Awen*0802^@tcp(localhost:3306)/gobatis?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		return
	}
	batis := gobatis.New(open)
	batis.Source("/")
	batis.ScanMappers(studentMapper, tag)
}

func TestAddOne(t *testing.T) {
	s := model.Student{
		Name:       "test",
		Age:        1,
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	if err = studentMapper.AddOne(s); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestAdds(t *testing.T) {
	var arr []any
	for i := 0; i < 10; i++ {
		s := model.Student{
			Name:       fmt.Sprintf("test_%d", i),
			Age:        i + 2,
			CreateTime: time.Now().Format("2006-01-02 15:04:05"),
		}
		arr = append(arr, s)
	}
	err = studentMapper.Adds(
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
	if stus, err = studentMapper.QueryAll(map[string]any{
		"table_name": "student",
		"id":         0,
	}); err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(stus)
}

func TestQueryPage(t *testing.T) {
	var stus []model.Student
	var count int64
	if stus, count, err = studentMapper.QueryPage(); err != nil {
		t.Error(err.Error())
		return
	}
	t.Log("rows:", stus, "count:", count)
}

func TestUpdate(t *testing.T) {
	var begin *sql.Tx
	var count int64
	begin, err = open.Begin()
	if err != nil {
		t.Error(err.Error())
		return
	}
	u := model.Student{
		Name: "awen",
		Age:  5,
	}
	count, err = studentMapper.Update(u, begin)
	if err != nil {
		t.Error(err.Error())
		return
	}
	begin.Commit()
	t.Log(count)
}

func TestIf(t *testing.T) {
	var stu model.Student
	args := map[string]any{
		"id":   1,
		"name": "test_0",
	}
	if stu, err = studentMapper.QueryIf(args); err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(stu)
}

func TestWhere(t *testing.T) {
	if err = tag.Where(); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestIsR(t *testing.T) {
	flag, err := studentMapper.IsRegister(model.Student{})
	if err != nil {
		t.Error(err.Error())
		return
	}
	fmt.Println(flag)
}

func TestNames(t *testing.T) {
	names, err := studentMapper.SelectNames()
	if err != nil {
		t.Error(err.Error())
		return
	}
	fmt.Println(names)
}

func TestName(t *testing.T) {
	t.Log(strconv.ParseBool("False"))
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

func TestMap(t *testing.T) {
}
