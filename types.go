package dagpb

import (
	ipld "github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/schema"
)

var Type struct {
	PBNode ipld.NodePrototype
	PBLink ipld.NodePrototype
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

var ts schema.TypeSystem

func init() {
	ts.Init()

	ts.Accumulate(schema.SpawnString("String"))
	ts.Accumulate(schema.SpawnInt("Int"))
	ts.Accumulate(schema.SpawnLink("Link"))
	ts.Accumulate(schema.SpawnBytes("Bytes"))

	/*
		type PBLink struct {
			Hash Link
			Name optional String
			Tsize optional Int
		}
	*/

	ts.Accumulate(schema.SpawnStruct("PBLink",
		[]schema.StructField{
			schema.SpawnStructField("Hash", "Link", false, false),
			schema.SpawnStructField("Name", "String", true, false),
			schema.SpawnStructField("Tsize", "Int", true, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))
	ts.Accumulate(schema.SpawnList("PBLinks", "PBLink", false))

	/*
		type PBNode struct {
			Links [PBLink]
			Data optional Bytes
		}
	*/

	ts.Accumulate(schema.SpawnStruct("PBNode",
		[]schema.StructField{
			schema.SpawnStructField("Links", "PBLinks", false, false),
			schema.SpawnStructField("Data", "Bytes", true, false),
		},
		schema.SpawnStructRepresentationMap(nil),
	))

	// TODO: do this only for the tests?
	// if errs := ts.ValidateGraph(); errs != nil {
	// 	for _, err := range errs {
	// 		fmt.Printf("- %s\n", err)
	// 	}
	// 	os.Exit(1)
	// }

	Type.PBNode = Prototype((*PBNode)(nil), ts.TypeByName("PBNode"))
	Type.PBLink = Prototype((*PBLink)(nil), ts.TypeByName("PBLink"))
}
