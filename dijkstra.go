package main

import "fmt"

// FindOptimalPath returns the optimal path from the source to any node in the graph
// It uses a modified version of Dijkstra's shortest path algorithm
func FindOptimalPath(graph *WeightedGraph, origin string, maxDistance int) {

	visited := make(map[string]bool)
	heap := &Heap{}

	startNode := graph.GetNode(origin)

	fmt.Println(startNode)

	startNode.value = 0
	heap.Push(startNode)

	for heap.Size() > 0 {
		current := heap.Pop()
		visited[current.name] = true

		edges := graph.Edges[current.name]

		for _, edge := range edges {
			// TODO: Add threshold condition below (current.value + edge.weight >= threshold, don't process further)
			if !visited[edge.node.name] {
				heap.Push(edge.node)

				if current.value+edge.weight < edge.node.value {
					// TODO: add the distance/people logic below.
					edge.node.value = current.value + edge.weight
					edge.node.through = current
				}
			}
		}
	}

}
