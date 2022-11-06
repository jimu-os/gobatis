package sqlgo

import "github.com/beevik/etree"

type Delete struct {
	Element  *etree.Element
	Fragment []*etree.Element
}

func (d *Delete) GetSql() *etree.Element {
	return d.Element
}

func (d *Delete) GetFragment() []*etree.Element {
	return d.Fragment
}
