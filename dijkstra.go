package main

import "fmt"

// FindOptimalPath returns the optimal path from the source to any node in the graph
// It uses a modified version of Dijkstra's shortest path algorithm
func FindOptimalPath(graph *WeightedGraph, origin string, destination string, maxDistance uint32, maxPassengers uint8) {
	logger.Log(fmt.Sprintf("--------------------\n"))
	logger.Log(fmt.Sprintf("Finding Optimal path\n"))
	logger.Log(fmt.Sprintf("--------------------\n"))

	visited := make(map[string]bool)
	heap := &Heap{}

	startNode := graph.GetNode(origin)

	startNode.shortestDistance = 0
	startNode.passengerCount = 1
	heap.Push(startNode)

	for heap.Size() > 0 {
		current := heap.Pop()
		visited[current.name] = true

		edges := graph.Edges[current.name]

		for _, edge := range edges {

			isVisited := visited[edge.node.name]
			hasCoveredMaxDistance := current.shortestDistance+edge.weight >= maxDistance
			hasReachedMaxPassengers := current.passengerCount > maxPassengers

			if !isVisited && !hasCoveredMaxDistance && !hasReachedMaxPassengers {
				heap.Push(edge.node)
				currentEmissionValue := (current.shortestDistance + edge.weight) / uint32(current.passengerCount+1)
				edgeNodeEmissionValue := edge.node.GetEmissionValue()

				if currentEmissionValue < edgeNodeEmissionValue {
					edge.node.shortestDistance = current.shortestDistance + edge.weight
					edge.node.through = current
					if edge.node.name != destination {
						edge.node.passengerCount = current.passengerCount + 1
					} else {
						edge.node.passengerCount = current.passengerCount
					}
				}
			}
		}
	}

}
