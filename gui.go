package main

import (
	"fmt"
	"strings"
)

func PrintChildren(n Node, depth int) {
	for i, c := range n.Children {
		if len(c.Children) > 0 {
			PrintChildren(*c, depth+1)
		}

		ind := strings.Repeat(" ", depth)
		if i == len(n.Children)-1 {
			fmt.Println(ind, "└──", c.URL.String())
		} else {
			fmt.Println(ind, "├──", c.URL.String())
		}
	}
}

func PrintMap(m Map) {
	// Print metadata
	count := CountChildrenOfNode(*m.Start)
	fmt.Println("Number of links found:", count)

	// Print tree
	fmt.Println("\n", m.Start.URL.String())
	PrintChildren(*m.Start, 0)
}