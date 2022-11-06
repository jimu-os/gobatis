package sqlgo

import "github.com/beevik/etree"

// Sql 单个xml的解析结构
type Sql struct {
	Element *etree.Element
	Selects map[string]Select
	Inserts map[string]Insert
	Updates map[string]Update
	Deletes map[string]Delete
}

func NewSql(root *etree.Element) *Sql {
	return &Sql{
		Element: root,
		Selects: map[string]Select{},
		Inserts: map[string]Insert{},
		Updates: map[string]Update{},
		Deletes: map[string]Delete{},
	}
}

func (receiver *Sql) LoadSqlElement() {
	elements := receiver.Element.ChildElements()
	for i := 0; i < len(elements); i++ {
		e := elements[i]
		switch e.Tag {
		case SELECT:
			s := Select{Element: e, Fragment: e.ChildElements()}
			receiver.Selects[e.SelectAttr("id").Value] = s
		case INSERT:
			ins := Insert{Element: e, Fragment: e.ChildElements()}
			receiver.Inserts[e.SelectAttr("id").Value] = ins
		case UPDATE:
			u := Update{Element: e, Fragment: e.ChildElements()}
			receiver.Updates[e.SelectAttr("id").Value] = u
		case DELETE:
			d := Delete{Element: e, Fragment: e.ChildElements()}
			receiver.Deletes[e.SelectAttr("id").Value] = d
		}
	}
}
