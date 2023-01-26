package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

const (
	HSFuldaUsername   = "HS"
	MaxPassengerCount = 2
	MaxDriverCount    = 2
)

var (
	HSFuldaCoordinates   = Location{50.565100, 9.686800}
	ActivePassengerCount = 0
	ActiveDriverCount    = 0
)

type Location struct {
	Lat, Long float64
}
type ConnectedNodesRequest struct {
	node  string
	value map[string]float32
}

const (
	Destination = iota
	Driver      = iota
	Passenger   = iota
)

type User struct {
	name     string
	location Location
	userType int
}

type GlobalState struct {
	activeUsers      []User
	activeUsersCh    chan User
	connectedNodes   map[string]map[string]float32
	connectedNodesCh chan ConnectedNodesRequest
}

// TODO: Ensure global state is not directly modified by any other goroutine
// Use channels for communication and modify from main goroutine
var globalData = GlobalState{
	activeUsers:      make([]User, 0, 100), // capacity of 100 users initially
	connectedNodes:   make(map[string]map[string]float32),
	connectedNodesCh: make(chan ConnectedNodesRequest),
	activeUsersCh:    make(chan User),
}

var quitCh = make(chan int)
var wg sync.WaitGroup

func main() {
	defer func() {
		wg.Wait()
		fmt.Printf("\nActive users:")
		fmt.Println(globalData.activeUsers)
		fmt.Printf("\nAvailable nodes:")
		fmt.Println(globalData.connectedNodes)
	}()

	// Don't wait initChannelListeners as it is in infinite for.
	go initChannelListeners(globalData.activeUsersCh, globalData.connectedNodesCh)
	addHSFuldaNode(globalData.activeUsersCh)

	// create N passengers and push to users channel  - concurrent
	// add listener to read from users channel and push to active users
	wg.Add(1)
	go addConcurrentPassengers(2, globalData.activeUsersCh)
	wg.Wait()
	wg.Add(1)
	go addConcurrentPassengers(2, globalData.activeUsersCh)
	wg.Wait()

	// create M drivers and push to users channel
	ActiveDriverCount += 1
	driverName := "d" + strconv.Itoa(ActiveDriverCount)
	wg.Add(1)
	go addDriver(driverName, globalData.activeUsersCh)
	wg.Wait()
	wg.Add(1)
	go assignPassengers(driverName, globalData.activeUsers, globalData.connectedNodes)
	wg.Wait()

}

func assignPassengers(driver string, users []User, connections map[string]map[string]float32) {
	graph := buildGraph(driver, users, connections)
	fmt.Printf("\nGraph")
	fmt.Println(graph)

	maxDistance := connections[driver][HSFuldaUsername] * 5 // TODO: find optimal multiplication factor
	FindOptimalPath(graph, driver, int(maxDistance))

	for _, node := range graph.Nodes {
		if node.name == HSFuldaUsername {
			fmt.Printf("Shortest path from %s to %s is %d\n", driver, HSFuldaUsername, node.value)
			for n := node; n.through != nil; n = n.through {
				fmt.Print(n.name, " <- ")

				// remove user (driver and passenger) from Users and Connections (don't remove HS fulda)
			}
			fmt.Println(driver)
			fmt.Println()
			break
		}
	}
	wg.Done()
}

func buildGraph(driver string, users []User, connections map[string]map[string]float32) *WeightedGraph {
	graph := NewGraph()
	nodes := graph.AddNodes(buildNodes(driver, users)...)

	fmt.Printf("\nNodes: ")
	fmt.Println(nodes)
	fmt.Println("Edges")

	for src, connection := range connections {
		for des, distance := range connection {
			fmt.Println(nodes[src], nodes[des], distance)
			// TODO: if multiple drivers causes issue, check this condition
			if nodes[src] != nil && nodes[des] != nil {
				graph.AddEdge(nodes[src], nodes[des], int(distance)) // TODO: update distance to float32
			}
		}
	}
	return graph
}

func buildNodes(driver string, users []User) (nodeNames []string) {
	for _, user := range users {
		if user.name == driver || user.userType != Driver {
			nodeNames = append(nodeNames, user.name)
		}
	}
	return
}

func initChannelListeners(usersCh chan User, nodesCh chan ConnectedNodesRequest) {
	var u User
	var cnr ConnectedNodesRequest

	for {
		select {
		case u = <-usersCh:
			wg.Add(1)
			go calculateDistanceToExistingNodes(u, globalData.activeUsers, nodesCh)
			globalData.activeUsers = append(globalData.activeUsers, u)

		case cnr = <-nodesCh:
			globalData.connectedNodes[cnr.node] = cnr.value

			// case <-quitCh:
			// 	wg.Done()
			// 	return

		}

	}
}

func calculateDistanceToExistingNodes(newUser User, existingUsers []User, nodesCh chan ConnectedNodesRequest) {
	var localWG sync.WaitGroup
	localData := make(map[string]float32)
	var mu sync.Mutex

	fmt.Println(len(existingUsers), newUser)

	for _, user := range existingUsers {
		if user.name == newUser.name {
			continue
		}

		localWG.Add(1)
		go func(user User) {
			distance := getDistance(newUser.location, user.location)
			// Use mutex to avoid concurrent writes to the map
			mu.Lock()
			localData[user.name] = distance
			mu.Unlock()
			localWG.Done()
		}(user)
	}
	localWG.Wait()

	fmt.Println(localData)
	nodesCh <- ConnectedNodesRequest{
		node:  newUser.name,
		value: localData,
	}

	wg.Done()
}

func addConcurrentPassengers(count int, ch chan<- User) {
	var localWG sync.WaitGroup
	for i := 0; i < count; i++ {
		ActivePassengerCount += 1
		localWG.Add(1)
		go func(i int) {
			ch <- User{
				name:     "p" + strconv.Itoa(i),
				location: generateRandomLocation(),
				userType: Passenger,
			}
			localWG.Done()
		}(ActivePassengerCount)
	}
	localWG.Wait()
	// quitCh <- 1
	wg.Done()
}

func addDriver(name string, ch chan<- User) {
	ch <- User{
		name:     name,
		location: generateRandomLocation(),
		userType: Driver,
	}
	wg.Done()
}

func getDistance(src Location, dest Location) float32 {
	// TODO: fetch data from OSRM
	time.Sleep(1000 * time.Millisecond)
	distance := float32(rand.Intn(100))
	return distance
}

// TODO: generate coordinates within a range around HS Fulda
func generateRandomLocation() Location {
	return Location{rand.Float64(), rand.Float64()}
}

func addHSFuldaNode(ch chan<- User) {
	ch <- User{
		name:     "HS",
		location: Location{50.565100, 9.686800},
		userType: Destination,
	}
}
