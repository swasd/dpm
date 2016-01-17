package build

import (
	"fmt"
	"testing"
)

func TestToposort(t *testing.T) {
	g := make(graph)
	g["a"] = []string{"b", "c"}
	g["c"] = []string{"b", "d"}
	g["b"] = []string{"d", "e"}
	g["d"] = []string{}
	g["e"] = []string{}
	order, cyclic := toposort(g)
	fmt.Println(order)
	fmt.Println(cyclic)
}

func TestToposortDependencies(t *testing.T) {

}
