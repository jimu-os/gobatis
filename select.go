package sqlgo

import "github.com/beevik/etree"

type Select struct {
	Element      *etree.Element
	ChildElement []*etree.Element
}
