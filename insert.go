package sqlgo

import "github.com/beevik/etree"

type Insert struct {
	Element  *etree.Element
	Fragment []*etree.Element
}

func (insert *Insert) GetSql() *etree.Element {
	return insert.Element
}

func (insert *Insert) GetFragment() []*etree.Element {
	return insert.Fragment
}
