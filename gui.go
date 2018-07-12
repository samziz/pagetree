package main

import (
	"os"
	"sort"
)

func PrintMap(m Map) {
	n := m.First

	for len(n.Children) > 0 {
		sortByAscendingChildrenLength(n.Children)
	}

	os.Stdout.Write()
}

func sortByAscendingChildrenLength(s *[]Node) {
	// Simple bubble sort to sort a Node's
	// children in descending order of length
	for i := 1; i < len(*s); i++ {
		if *s[i] > *s[i-1] {
			*s[i], *s[i-1] = *s[i-1], *s[i]
		}
	}
}