package main

import (
	"math"
	"sync"
)

type Node struct {
	name    string
	value   int
	through *Node // TODO: check if through can be used to find number of nodes at a point
}

type Edge struct {
	node   *Node
	weight int // TODO: CHANGE TYPE TO FLOAT32
}

type WeightedGraph struct {
	Nodes []*Node
	Edges map[string][]*Edge
	mutex sync.RWMutex
}

func NewGraph() *WeightedGraph {
	return &WeightedGraph{
		Edges: make(map[string][]*Edge),
	}
}

func (g *WeightedGraph) GetNode(name string) (node *Node) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	for _, n := range g.Nodes {
		if n.name == name {
			node = n // TODO: return early?
		}
	}
	return
}

func (g *WeightedGraph) AddNode(n *Node) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.Nodes = append(g.Nodes, n)
}

func (graph *WeightedGraph) AddNodes(names ...string) (nodes map[string]*Node) {
	nodes = make(map[string]*Node)

	for _, name := range names {
		n := &Node{name: name, value: math.MaxInt, through: nil}
		graph.AddNode(n)
		nodes[name] = n
	}
	return
}

func (g *WeightedGraph) AddEdge(n1, n2 *Node, weight int) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.Edges[n1.name] = append(g.Edges[n1.name], &Edge{node: n2, weight: weight})
	g.Edges[n2.name] = append(g.Edges[n2.name], &Edge{node: n1, weight: weight})
}
