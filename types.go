package dagpb

import (
	ipld "github.com/ipld/go-ipld-prime"
)

var Type = struct {
	PBNode ipld.NodePrototype
	PBLink ipld.NodePrototype
}{
	PBNode: Prototype((*PBNode)(nil)),
	PBLink: Prototype((*PBLink)(nil)),
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
