package sqlgo

import "github.com/beevik/etree"

type Update struct {
	Element  *etree.Element
	Fragment []*etree.Element
}
