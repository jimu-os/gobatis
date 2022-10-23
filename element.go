package sqlgo

import "github.com/beevik/etree"

type Insert struct {
	Element      *etree.Element
	ChildElement []*etree.Element
}

type Update struct {
	Element      *etree.Element
	ChildElement []*etree.Element
}

type Delete struct {
	Element      *etree.Element
	ChildElement []*etree.Element
}

type For struct {
	Element *etree.Element
}

type If struct {
	Element *etree.Element
}
