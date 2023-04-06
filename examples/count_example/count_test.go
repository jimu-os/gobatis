package count_example

import (
	"database/sql"
	"gitee.com/aurora-engine/gobatis"
	_ "github.com/go-sql-driver/mysql"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

var countMapper *CountMapper

func init() {
	countMapper = &CountMapper{}
	open, err := sql.Open("mysql", "root:Aurora@2022@tcp(82.157.160.117:3306)/community?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		return
	}
	batis := gobatis.New(open)
	batis.Source("/")
	batis.ScanMappers(countMapper)
}

func TestCount(t *testing.T) {
	students, count, err := countMapper.Select(1, 2)
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(students)
	t.Log(count)
}

var srcCode = `
package count_example

import "gitee.com/aurora-engine/gobatis/examples/count_example/model"

type CountMapper struct {
	// @select select * from student limit 10 offset 0
	Select func() ([]model.Student, int64, error)
	Insert func(id string, name string)
}
`

func TestAST(t *testing.T) {
	fileSet := token.NewFileSet()
	//file, err := parser.ParseFile(fileSet, "", srcCode, parser.ParseComments)
	file, err := parser.ParseFile(fileSet, "", CountMapper{}, parser.ParseComments)
	if err != nil {
		t.Error(err.Error())
		return
	}
	ast.Print(fileSet, file)
}
