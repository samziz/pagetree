package main

import (
	"net/url"
	)

// Map is a standard tree structure containing the first
// Node, i.e. the URL the user starts crawling from.
type Map struct {
	Start *Node
}

type Node struct {
	Children []*Node
	URL url.URL
}


func CountChildrenOfNode(n Node) int {
	count := 1

	for _, c := range n.Children {
		count += CountChildrenOfNode(*c)
	}

	return count
}

func CreateNodesFromURLs(urls []url.URL) []Node {
	nodes := make([]Node, len(urls))

	for i, u := range urls {
		nodes[i] = Node{URL: u}
	}

	return nodes
}