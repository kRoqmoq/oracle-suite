package main

import (
	"fmt"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider/graph/nodes"
)

func main() {
	g := nodes.NewMedianAggregatorNode(provider.Pair{Base: "ETH", Quote: "USD"}, 1)
	o := nodes.NewOriginNode(nodes.OriginPair{
		Origin: "test",
		Pair:   provider.Pair{Base: "ETH", Quote: "USD"},
	}, 0, 0)

	g.AddChild(o)
	//f := NewFeeder(s, null.New())
	//warns := f.Feed([]nodes.Node{g}, time.Now())
	fmt.Printf("%v", g)
}
