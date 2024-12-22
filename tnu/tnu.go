package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cosi-project/runtime/pkg/resource"
	"github.com/distribution/reference"
	"github.com/siderolabs/talos/pkg/machinery/client"
	"github.com/siderolabs/talos/pkg/machinery/resources/config"
)

func main() {
	ctx := context.Background()
	if len(os.Args) < 2 || len(os.Args[1]) == 0 {
		log.Fatal("usage: tnu <tag>")
	}
	tag := os.Args[1]
	c, err := client.New(ctx, client.WithDefaultConfig())
	if err != nil {
		panic(err)
	}
	var rns resource.Namespace
	rd, err := c.ResolveResourceKind(ctx, &rns, "MachineConfig")
	if err != nil {
		panic(err)
	}
	r, err := c.COSI.Get(ctx, resource.NewMetadata(rns, rd.TypedSpec().Type, "v1alpha1", resource.VersionUndefined))
	if err != nil {
		panic(err)
	}
	mc, ok := r.(*config.MachineConfig)
	if !ok {
		panic("not a config.MachineConfig")
	}
	image := mc.Config().Machine().Install().Image()
	ref, err := reference.ParseAnyReference(image)
	if err != nil {
		panic(err)
	}
	ntref, ok := ref.(reference.NamedTagged)
	if !ok {
		panic("not a reference.NamedTagged")
	}
	v, err := c.Version(ctx)
	if err != nil {
		panic(err)
	}
	if v.Messages[0].GetVersion().GetTag() == tag {
		log.Printf("version is already %s", tag)
		os.Exit(0)
	}
	ntref, err = reference.WithTag(ntref, tag)
	if err != nil {
		panic(err)
	}
	log.Printf("upgrading to %s", ntref)
	resp, err := c.UpgradeWithOptions(ctx,
		client.WithUpgradeImage(ntref.String()),
		client.WithUpgradePreserve(true),
		client.WithUpgradeStage(true),
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("upgrade started: %s\n", resp.GetMessages()[0].String())
}
