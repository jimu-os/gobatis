package sgo

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/beevik/etree"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

func New(db *sql.DB) *Build {
	return &Build{
		DB:         db,
		NameSpaces: map[string]*Sql{},
	}
}

type Build struct {
	DB        *sql.DB
	SqlSource string
	// 保存所有的 xml 解析
	NameSpaces map[string]*Sql
}

func (build *Build) Source(source string) {
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
			element := document.Root()
			attr := element.SelectAttr("namespace")
			s := NewSql(element)
			s.LoadSqlElement()
			build.NameSpaces[attr.Value] = s
		}
		return nil
	})
}

// ScanMappers 扫描解析
func (build *Build) ScanMappers(mappers ...any) {
	for i := 0; i < len(mappers); i++ {
		mapper := mappers[i]
		vf := reflect.ValueOf(mapper)
		if vf.Kind() != reflect.Pointer {
			panic("")
		}
		if vf.Elem().Kind() != reflect.Struct {
			panic("")
		}
		vf = vf.Elem()
		namespace := vf.Type().String()
		namespace = Namespace(namespace)
		for j := 0; j < vf.NumField(); j++ {
			key := make([]string, 0)
			key = append(key, namespace)
			structField := vf.Type().Field(i)
			field := vf.Field(j)
			if !structField.IsExported() || structField.Type.Kind() != reflect.Func {
				continue
			}
			key = append(key, structField.Name)
			build.initMapper(key, field)
		}
	}
}

func (build *Build) Sql(id string, value any) (string, error) {
	ids := strings.Split(id, ".")
	if len(ids) != 2 {
		return "", errors.New("id error")
	}
	ctx := toMap(value)
	if sql, b := build.NameSpaces[ids[0]]; b {
		if element, f := sql.Statement[ids[1]]; f {
			analysis, _, err := Analysis(element, ctx)
			if err != nil {
				return "", err
			}
			join := strings.Join(analysis, " ")
			return join, nil
		}
	}
	return "", nil
}

func (build *Build) Get(id []string, value any) (string, string, error) {
	if len(id) != 2 {
		return "", "", errors.New("id error")
	}
	ctx := toMap(value)
	if sql, b := build.NameSpaces[id[0]]; b {
		if element, f := sql.Statement[id[1]]; f {
			analysis, tag, err := Analysis(element, ctx)
			if err != nil {
				return "", "", err
			}
			join := strings.Join(analysis, " ")
			return join, tag, nil
		}
	}
	return "", "", fmt.Errorf("not found sql statement element")
}

// Analysis 解析xml标签
func Analysis(element *etree.Element, ctx map[string]any) ([]string, string, error) {
	var err error
	sql := []string{}
	// 解析根标签 开始之后的文本
	sqlStar := element.Text()
	// 处理字符串前后空格
	sqlStar = strings.TrimSpace(sqlStar)
	//更具标签类型，对应解析字符串
	sqlStar, err = Element(element, sqlStar, ctx)
	if err != nil {
		return nil, "", err
	}
	sql = append(sql, sqlStar)
	// if 标签解析 逻辑不通过
	if sqlStar != "" && err == nil {
		// 解析子标签内容
		child := element.ChildElements()
		for _, childElement := range child {
			analysis, _, err := Analysis(childElement, ctx)
			if err != nil {
				return nil, "", fmt.Errorf("%s -> %s error,%s", element.Tag, childElement.Tag, err.Error())
			}
			sql = append(sql, analysis...)
		}
	}
	endSql := element.Tail()
	endSql = strings.TrimSpace(endSql)
	if endSql != "" {
		endSql, err = Element(element.Parent(), endSql, ctx)
		if err != nil {
			return nil, "", err
		}
		sql = append(sql, endSql)
	}
	return sql, element.Tag, nil
}

func Element(element *etree.Element, template string, ctx map[string]any) (string, error) {
	// 检擦 节点标签类型
	tag := element.Tag
	switch tag {
	case For:
		return ForElement(element, template, ctx)
	case If:
		return IfElement(element, template, ctx)
	case Select, Update, Delete, Insert:
		return StatementElement(element, template, ctx)
	case Mapper:
		// 对根标签不做任何处理
		return "", nil
	}
	return "", errors.New("error")
}

func Namespace(namespace string) string {
	if index := strings.LastIndex(namespace, "."); index != -1 {
		return namespace[index+1:]
	}
	return namespace
}
