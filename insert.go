package sqlgo

import "github.com/beevik/etree"

type Insert struct {
	Element  *etree.Element
	Fragment []*etree.Element
}
