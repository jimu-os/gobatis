package sqlgo

import (
	"github.com/beevik/etree"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	SELECT = "select"
	INSERT = "insert"
	UPDATE = "update"
	DELETE = "delete"
)

type Build struct {
	SqlSource string
	// 保存所有的 xml 解析
	NameSpaces map[string]*Sql
}

func (build *Build) LoadXml(source string) {
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

func Analysis(element *etree.Element) (string, error) {

	return "", nil
}
