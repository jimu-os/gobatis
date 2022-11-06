package sqlgo

import "github.com/beevik/etree"

type Statement interface {
	GetSql() *etree.Element
}

type Fragment interface {
	GetFragment() []*etree.Element
}

type For struct {
	Element *etree.Element
}

type If struct {
	Element  *etree.Element
	Fragment []*etree.Element
}
