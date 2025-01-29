package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cosi-project/runtime/pkg/resource"
	"github.com/distribution/reference"
	"github.com/siderolabs/talos/pkg/machinery/client"
	"github.com/siderolabs/talos/pkg/machinery/resources/config"
	"github.com/siderolabs/talos/pkg/machinery/resources/k8s"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	if len(os.Args) < 3 || len(os.Args[1]) == 0 || len(os.Args[2]) == 0 {
		log.Fatal("usage: tnu <node> <tag>")
	}
	nodename := os.Args[1]
	tag := os.Args[2]
	ctx := client.WithNode(context.Background(), nodename)
	c, err := client.New(ctx, client.WithDefaultConfig())
	if err != nil {
		panic(err)
	}

	r, err := c.COSI.Get(ctx, resource.NewMetadata("k8s", "Nodenames.kubernetes.talos.dev", "nodename", resource.VersionUndefined))
	if err != nil {
		panic(err)
	}
	nn, ok := r.(*k8s.Nodename)
	if !ok {
		panic("not a k8s.Nodename")
	}
	if nn.TypedSpec().Nodename != nodename {
		log.Fatalf("expected node %s, got %s", nodename, nn.TypedSpec().Nodename)
	}

	r, err = c.COSI.Get(ctx, resource.NewMetadata("config", "MachineConfigs.config.talos.dev", "v1alpha1", resource.VersionUndefined))
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

	vresp, err := c.Version(ctx)
	if err != nil {
		panic(err)
	}
	mcSchematic := ntref.Name()[strings.LastIndex(ntref.Name(), "/")+1:]
	nodeSchematic := getSchematicAnnotation(ctx, nodename)
	if vresp.Messages[0].GetVersion().GetTag() == tag && mcSchematic == nodeSchematic {
		log.Printf("node is up-to-date (schematic: %s, tag: %s)", mcSchematic, tag)
		os.Exit(0)
	}

	ntref, err = reference.WithTag(ntref, tag)
	if err != nil {
		panic(err)
	}

	log.Printf("upgrading %s to %s", nodename, ntref)
	uresp, err := c.UpgradeWithOptions(ctx,
		client.WithUpgradeImage(ntref.String()),
		client.WithUpgradePreserve(true),
		client.WithUpgradeStage(true),
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("upgrade started: %s\n", uresp.GetMessages()[0].String())
	<-make(chan int, 1)
}

func getSchematicAnnotation(ctx context.Context, nodename string) string {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	node, err := clientset.CoreV1().Nodes().Get(ctx, nodename, metav1.GetOptions{})
	if err != nil {
		panic(err)
	}
	v, ok := node.Annotations["extensions.talos.dev/schematic"]
	if !ok {
		panic("extensions.talos.dev/schematic annotation not found")
	}
	return v
}
