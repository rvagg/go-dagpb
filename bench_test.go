package dagpb_test

import (
	"bytes"
	"fmt"
	"testing"

	cid "github.com/ipfs/go-cid"
	dagpb "github.com/ipld/go-codec-dagpb"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/fluent/qp"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
)

var benchInput []byte

func init() {
	someData := bytes.Repeat([]byte("some plaintext data\n"), 10)
	someCid, _ := cid.Cast([]byte{1, 85, 0, 5, 0, 1, 2, 3, 4})

	node, err := qp.BuildMap(dagpb.Type.PBNode, 2, func(ma ipld.MapAssembler) {
		qp.MapEntry(ma, "Data", qp.Bytes(someData))
		qp.MapEntry(ma, "Links", qp.List(10, func(la ipld.ListAssembler) {
			for i := 0; i < 10; i++ {
				qp.ListEntry(la, qp.Map(3, func(ma ipld.MapAssembler) {
					qp.MapEntry(ma, "Hash", qp.Link(cidlink.Link{Cid: someCid}))
					qp.MapEntry(ma, "Name", qp.String(fmt.Sprintf("%d", i)))
					qp.MapEntry(ma, "Tsize", qp.Int(10))
				}))
			}
		}))
	})
	if err != nil {
		panic(err)
	}
	enc, err := dagpb.AppendEncode(nil, node)
	if err != nil {
		panic(err)
	}
	benchInput = enc
	// println(len(benchInput))
}

func BenchmarkRoundtrip(b *testing.B) {
	var enc []byte
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			nb := dagpb.Type.PBNode.NewBuilder()
			if err := dagpb.DecodeBytes(nb, benchInput); err != nil {
				b.Fatal(err)
			}
			node := nb.Build()

			enc = enc[:0] // reuse the buffer

			enc, err := dagpb.AppendEncode(enc, node)
			if err != nil {
				b.Fatal(err)
			}
			// println(len(benchInput), len(enc))
			_ = enc
		}
	})
}
