package sqlgo

import "github.com/beevik/etree"

type Select struct {
	Element  *etree.Element
	Fragment []*etree.Element
}

func (s *Select) GetSql() *etree.Element {
	return s.Element
}

func (s *Select) GetFragment() []*etree.Element {
	return s.Fragment
}
