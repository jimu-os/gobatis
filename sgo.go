package sqlgo

import (
	"errors"
	"github.com/beevik/etree"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func NewSqlGo() *Sgo {
	return &Sgo{
		NameSpaces: map[string]*Sql{},
	}
}

type Sgo struct {
	SqlSource string
	// 保存所有的 xml 解析
	NameSpaces map[string]*Sql
}

func (build *Sgo) LoadXml(source string) {
	if source != "" {
		build.SqlSource = source
	}
	getwd, err := os.Getwd()
	if err != nil {
		return
	}
	root := filepath.Join(getwd, build.SqlSource)

	// 解析 xml
	filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if strings.HasSuffix(path, ".xml") {
			document := etree.NewDocument()
			err = document.ReadFromFile(path)
			if err != nil {
				return err
			}
			element := document.SelectElement("sql")
			attr := element.SelectAttr("namespace")
			s := NewSql(element)
			s.LoadSqlElement()
			build.NameSpaces[attr.Value] = s
		}
		return nil
	})
}

func (build *Sgo) Sql(id string, ctx map[string]any) (string, error) {
	ids := strings.Split(id, ".")
	if len(ids) != 2 {
		return "", errors.New("id error")
	}
	if sql, b := build.NameSpaces[ids[0]]; b {
		if element, f := sql.Statement[ids[1]]; f {
			analysis, err := Analysis(element, ctx)
			if err != nil {
				return "", err
			}
			join := strings.Join(analysis[:len(analysis)-1], " ")
			return join, nil
		}
	}
	return "", nil
}

// Analysis 解析xml标签，sql root (select，insert，update，delete)，不解析开始标签之前，结束标签之后的文本内容
func Analysis(element *etree.Element, ctx map[string]any) ([]string, error) {
	sql := []string{}
	// 解析根标签 开始之后的文本
	sqlStar := element.Text()
	// 处理字符串前后空格
	sqlStar = strings.TrimSpace(sqlStar)
	//更具标签类型，对应解析字符串
	sqlStar, err := Element(element, sqlStar, ctx)
	if err != nil {
		return nil, err
	}
	sql = append(sql, sqlStar)
	// 解析子标签内容
	child := element.ChildElements()
	for _, element := range child {
		analysis, err := Analysis(element, ctx)
		if err != nil {
			return nil, err
		}
		sql = append(sql, analysis...)
	}
	endSql := element.Tail()
	endSql = strings.TrimSpace(endSql)
	if endSql != "" {
		endSql, err = Element(element.Parent(), endSql, ctx)
		if err != nil {
			return nil, err
		}
		sql = append(sql, endSql)
	}
	return sql, nil
}

func Element(element *etree.Element, template string, ctx map[string]any) (string, error) {
	// 检擦 节点标签类型
	tag := element.Tag
	switch tag {
	case "for":
		return ForElement(element, template, ctx)
	case "if":
		return IfElement(element, template, ctx)
	case "select", "update", "delete", "insert":
		return StatementElement(element, template, ctx)
	}
	return "", errors.New("error")
}
