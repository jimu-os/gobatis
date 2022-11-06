package sqlgo

import "github.com/beevik/etree"

type Update struct {
	Element  *etree.Element
	Fragment []*etree.Element
}

func (u *Update) GetSql() *etree.Element {
	return u.Element
}

func (u *Update) GetFragment() []*etree.Element {
	return u.Fragment
}
