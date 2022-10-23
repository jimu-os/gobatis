package sqlgo

import "github.com/beevik/etree"

type Sql struct {
	Element *etree.Element
	Selects map[string]*Select
	Inserts map[string]*Insert
	Updates map[string]*Update
	Deletes map[string]*Delete
}

func NewSql(root *etree.Element) *Sql {
	return &Sql{
		Element: root,
		Selects: map[string]*Select{},
		Inserts: map[string]*Insert{},
		Updates: map[string]*Update{},
		Deletes: map[string]*Delete{},
	}
}

func (receiver *Sql) LoadSqlElement() {
	elements := receiver.Element.ChildElements()
	for i := 0; i < len(elements); i++ {
		e := elements[i]
		switch e.Tag {
		case "select":
			s := &Select{Element: e, ChildElement: e.ChildElements()}
			receiver.Selects[e.SelectAttr("id").Value] = s
		case "insert":
			ins := &Insert{Element: e, ChildElement: e.ChildElements()}
			receiver.Inserts[e.SelectAttr("id").Value] = ins
		case "update":
			u := &Update{Element: e, ChildElement: e.ChildElements()}
			receiver.Updates[e.SelectAttr("id").Value] = u
		case "delete":
			d := &Delete{Element: e, ChildElement: e.ChildElements()}
			receiver.Deletes[e.SelectAttr("id").Value] = d
		}
	}
}
