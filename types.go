package dagpb

import (
	"reflect"

	ipld "github.com/ipld/go-ipld-prime"
)

var Type struct {
	PBNode ipld.NodePrototype
	PBLink ipld.NodePrototype
}

func init() {
	Type.PBNode = _prototype{reflect.TypeOf(PBNode{})}
	Type.PBLink = _prototype{reflect.TypeOf(PBLink{})}
}

type PBNode struct {
	Links []PBLink
	Data  *[]byte // optional
}

type PBLink struct {
	Hash  ipld.Link
	Name  *string // optional
	Tsize *int    // optional
}
