package sqlgo

import "github.com/beevik/etree"

type Delete struct {
	Element  *etree.Element
	Fragment []*etree.Element
}
