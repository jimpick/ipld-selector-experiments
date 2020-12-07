package main

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	_ "github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/ipld/go-ipld-prime/fluent"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	"github.com/ipld/go-ipld-prime/traversal"
	"github.com/ipld/go-ipld-prime/traversal/selector/builder"
)

var linkBuilder = cidlink.LinkBuilder{cid.Prefix{
	Version:  1,    // Usually '1'.
	Codec:    0x71, // dag-cbor as per multicodec
	MhType:   0x15, // sha3-384 as per multihash
	MhLength: 48,   // sha3-384 hash has a 48-byte sum.
}}

func main() {
	ctx := context.Background()
	storage := make(map[ipld.Link][]byte)
	store := func(ipld.LinkContext) (io.Writer, ipld.StoreCommitter, error) {
		buf := bytes.Buffer{}
		return &buf, func(lnk ipld.Link) error {
			storage[lnk] = buf.Bytes()
			return nil
		}, nil
	}
	loader := func(lnk ipld.Link, _ ipld.LinkContext) (io.Reader, error) {
		return bytes.NewReader(storage[lnk]), nil
	}

	eric := fluent.MustBuildMap(basicnode.Prototype.Any, 1, func(na fluent.MapAssembler) {
		na.AssembleEntry("name").AssignString("Eric Myhre")
	})
	ericLink, _ := linkBuilder.Build(ctx, ipld.LinkContext{}, eric, store)
	daniel := fluent.MustBuildMap(basicnode.Prototype.Any, 1, func(na fluent.MapAssembler) {
		na.AssembleEntry("name").AssignString("Daniel Martí")
	})
	danielLink, _ := linkBuilder.Build(ctx, ipld.LinkContext{}, daniel, store)
	jim := fluent.MustBuildMap(basicnode.Prototype.Any, 1, func(na fluent.MapAssembler) {
		na.AssembleEntry("name").AssignString("Jim Pick")
	})
	jimLink, _ := linkBuilder.Build(ctx, ipld.LinkContext{}, jim, store)
	people := fluent.MustBuildList(basicnode.Prototype.Any, 2, func(na fluent.ListAssembler) {
		na.AssembleValue().AssignLink(ericLink)
		na.AssembleValue().AssignLink(danielLink)
		na.AssembleValue().AssignLink(jimLink)
	})
	peopleLink, _ := linkBuilder.Build(ctx, ipld.LinkContext{}, people, store)

	nb := basicnode.Prototype.Any.NewBuilder()
	_ = peopleLink.Load(ctx, ipld.LinkContext{}, nb, loader)
	people2 := nb.Build()
	for itr := people2.ListIterator(); !itr.Done(); {
		_, value, _ := itr.Next()
		personLink, _ := value.AsLink()

		nb := basicnode.Prototype.Any.NewBuilder()
		_ = personLink.Load(ctx, ipld.LinkContext{}, nb, loader)
		person := nb.Build()

		name, _ := person.LookupByString("name")
		nameStr, _ := name.AsString()
		fmt.Println(nameStr)
	}

	ssb := builder.NewSelectorSpecBuilder(basicnode.Prototype.Any)

	/*
		partialSelector := ssb.ExploreFields(func(specBuilder builder.ExploreFieldsSpecBuilder) {
			specBuilder.Insert("Links", ssb.ExploreIndex(0, ssb.ExploreFields(func(specBuilder builder.ExploreFieldsSpecBuilder) {
				specBuilder.Insert("Hash", ssb.Matcher())
			})))
		}).Node()
	*/
	ss := ssb.ExploreRange(2, 3, ssb.Matcher())
	s, err := ss.Selector()
	if err != nil {
		panic(err)
	}
	err = traversal.Progress{
		Cfg: &traversal.Config{
			LinkLoader: func(lnk ipld.Link, _ ipld.LinkContext) (io.Reader, error) {
				return bytes.NewReader(storage[lnk]), nil
			},
			LinkTargetNodePrototypeChooser: func(_ ipld.Link, _ ipld.LinkContext) (ipld.NodePrototype, error) {
				return basicnode.Prototype__Any{}, nil
			},
		},
	}.WalkMatching(people, s, func(prog traversal.Progress, n ipld.Node) error {
		name, _ := n.LookupByString("name")
		nameStr, _ := name.AsString()
		fmt.Println("Match:", nameStr)
		return nil
	})
	if err != nil {
		panic(err)
	}

}
