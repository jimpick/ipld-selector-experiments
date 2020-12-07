package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	_ "github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/ipld/go-ipld-prime/fluent"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
)

var linkBuilder = cidlink.LinkBuilder{cid.Prefix{
	Version:  1,    // Usually '1'.
	Codec:    0x71, // dag-cbor as per multicodec
	MhType:   0x15, // sha3-384 as per multihash
	MhLength: 48,   // sha3-384 hash has a 48-byte sum.
}}

func main() {
	ctx := context.Background()
	// For now, we're not storing the blocks anywhere.
	store := func(ipld.LinkContext) (io.Writer, ipld.StoreCommitter, error) {
		return ioutil.Discard, func(lnk ipld.Link) error { return nil }, nil
	}

	eric := fluent.MustBuildMap(basicnode.Prototype.Any, 1, func(na fluent.MapAssembler) {
		na.AssembleEntry("name").AssignString("Eric Myhre")
	})
	ericLink, _ := linkBuilder.Build(ctx, ipld.LinkContext{}, eric, store)
	people := fluent.MustBuildList(basicnode.Prototype.Any, 1, func(na fluent.ListAssembler) {
		na.AssembleValue().AssignLink(ericLink)
	})
	peopleLink, _ := linkBuilder.Build(ctx, ipld.LinkContext{}, people, store)
	fmt.Println(peopleLink)
}
