package sqlgo

import "github.com/beevik/etree"

// Sql 单个xml的解析结构
type Sql struct {
	Element   *etree.Element
	Statement map[string]*etree.Element
}

func NewSql(root *etree.Element) *Sql {
	return &Sql{Element: root, Statement: map[string]*etree.Element{}}
}

func (receiver *Sql) LoadSqlElement() {
	elements := receiver.Element.ChildElements()
	for i := 0; i < len(elements); i++ {
		e := elements[i]
		key := e.SelectAttr("id").Value
		receiver.Statement[key] = e
	}
}
