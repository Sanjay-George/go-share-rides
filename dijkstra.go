package main

import "fmt"

// FindOptimalPath returns the optimal path from the source to any node in the graph
// It uses a modified version of Dijkstra's shortest path algorithm
func FindOptimalPath(graph *WeightedGraph, origin string, destination string, maxDistance uint32) {
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
			// TODO: Add threshold condition below (current.value + edge.weight >= threshold, don't process further)
			// fmt.Printf("currentNode: %s edgeNode: %s\n", current.name, edge.node.name)
			// fmt.Printf("current.shortestDistance: %d, edge.weight: %d \n", current.shortestDistance, edge.weight)

			if !visited[edge.node.name] && !(current.shortestDistance+edge.weight >= maxDistance) {
				heap.Push(edge.node)
				currentEmissionValue := (current.shortestDistance + edge.weight) / uint32(current.passengerCount+1)
				edgeNodeEmissionValue := edge.node.GetEmissionValue()
				// fmt.Printf("currentEmissionValue: %d, edge.node.shortestDistance: %d, current.passengerCount: %d, edgeEmissionValue: %d \n", currentEmissionValue, edge.node.shortestDistance, current.passengerCount, edgeNodeEmissionValue)

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
