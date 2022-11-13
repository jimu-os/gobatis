package sgo

import "github.com/beevik/etree"

const (
	Select = "select"
	Insert = "insert"
	Update = "update"
	Delete = "delete"
	Mapper = "mapper"
	For    = "for"
	If     = "if"
)

// Sql 单个xml的解析结构
type Sql struct {
	// Element 表示 一个 Mapper 文件的更元素
	Element *etree.Element
	// Statement 表示每个 更元素下面的 sql语句标签
	Statement map[string]*etree.Element
}

func NewSql(root *etree.Element) *Sql {
	return &Sql{Element: root, Statement: map[string]*etree.Element{}}
}

func (receiver *Sql) LoadSqlElement() {
	elements := receiver.Element.ChildElements()
	for i := 0; i < len(elements); i++ {
		e := elements[i]
		key := e.SelectAttr("id")
		if key != nil {
			receiver.Statement[key.Value] = e
		}
	}
}
