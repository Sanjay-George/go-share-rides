package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"
)

const (
	HSFuldaUsername      = "HS"
	MaxPassengerCount    = 2
	MaxDriverCount       = 2
	MultiplicationFactor = 1.5
)

var (
	HSFuldaCoordinates   = Location{50.565100, 9.686800}
	ActivePassengerCount = 0
	ActiveDriverCount    = 0
	isBenchMarked        = false
)

// TODO: Ensure global state is not directly modified by any other goroutine
// Use channels for communication and modify from main goroutine
var globalData = GlobalState{
	activeUsers:      make([]User, 0, 100), // capacity of 100 users initially
	connectedNodes:   make(map[string]map[string]int),
	connectedNodesCh: make(chan ConnectedNodesRequest),
	activeUsersCh:    make(chan User),
}

// var quitCh = make(chan int)
var wg sync.WaitGroup
var rwMutex sync.RWMutex
var logger Logger

func main() {
	start := time.Now()

	defer func(start time.Time) {
		wg.Wait()
		logger.Log(fmt.Sprintf("-------------\n"))
		logger.Log(fmt.Sprintf("Active users:\n"))
		logger.Log(fmt.Sprintf("-------------\n"))
		logger.Log(fmt.Sprintln(globalData.activeUsers))
		logger.Log(fmt.Sprintf("-------------\n"))
		logger.Log(fmt.Sprintf("Available nodes:\n"))
		logger.Log(fmt.Sprintf("-------------\n"))
		logger.Log(fmt.Sprintln(globalData.connectedNodes))
		duration := time.Since(start)
		fmt.Printf("\n----------------------\n")
		fmt.Printf("Execution Time: %d ms\n", duration.Milliseconds())
		fmt.Printf("----------------------\n")
	}(start)

	rand.Seed(time.Now().UnixNano())
	isBenchMarked = false
	if !isBenchMarked {
		logger.isEnabled = true
	}

	// runtime.GOMAXPROCS(1)
	logger.Log(fmt.Sprintln("----------------------------------------------------------"))
	logger.Log(fmt.Sprintf("Welcome to Carida! Threads: %d. Available CPU: %d\n", runtime.GOMAXPROCS(-1), runtime.NumCPU()))
	logger.Log(fmt.Sprintln("----------------------------------------------------------"))

	// Don't wait initChannelListeners as it is in infinite for.
	go initChannelListeners(globalData.activeUsersCh, globalData.connectedNodesCh)
	addHSFuldaNode(globalData.activeUsersCh)

	wg.Add(1)
	go addConcurrentPassengers(50, globalData.activeUsersCh)
	wg.Wait()

	wg.Add(1)
	go addConcurrentDrivers(10, globalData.activeUsersCh)
	wg.Wait()

	wg.Add(1)
	go assignPassengersToActiveDrivers(globalData.activeUsers, globalData.connectedNodes)
	wg.Wait()

}

func addConcurrentDrivers(count int, ch chan<- User) {
	var localWG sync.WaitGroup

	for i := 0; i < count; i++ {
		ActiveDriverCount += 1
		driverName := "d" + strconv.Itoa(ActiveDriverCount)
		localWG.Add(1)
		go func(name string) {
			ch <- User{
				name:     name,
				location: generateRandomLocation(30),
				userType: Driver,
			}
			localWG.Done()
		}(driverName)
	}
	localWG.Wait()
	wg.Done()
}

func assignPassengersToActiveDrivers(users []User, connections map[string]map[string]int) {
	var activeDrivers []string
	var localWG sync.WaitGroup

	// Get all active drivers
	rwMutex.RLock()
	for _, user := range users {
		if user.userType == Driver {
			activeDrivers = append(activeDrivers, user.name)
		}
	}
	rwMutex.RUnlock()
	logger.Log(fmt.Sprintln("Active Drivers"))
	logger.Log(fmt.Sprintln(activeDrivers))

	// for each driver, call assignPassengers goroutine
	for _, driver := range activeDrivers {
		localWG.Add(1)
		go assignPassengers(driver, users, connections, &localWG)
	}
	localWG.Wait()
	wg.Done()
}

func assignPassengers(driver string, users []User, connections map[string]map[string]int, parentWG *sync.WaitGroup) {
	graph := buildGraph(driver, users, connections)

	maxDistance := int(float32(connections[driver][HSFuldaUsername]) * MultiplicationFactor) // TODO: find optimal multiplication factor
	FindOptimalPath(graph, driver, HSFuldaUsername, int(maxDistance))

	for _, node := range graph.Nodes {
		if node.name != HSFuldaUsername {
			continue
		}
		if node.shortestDistance == math.MaxInt {
			logger.Log(fmt.Sprintf("No optimal path found from %s to %s\n", driver, HSFuldaUsername))
			break
		}
		logger.Log(fmt.Sprintf("Optimal path from %s to %s covers %d m with emission of %d units per person\n", driver, HSFuldaUsername, node.shortestDistance, node.GetEmissionValue()))
		for n := node; n.through != nil; n = n.through {
			logger.Log(fmt.Sprintf("%s <- ", n.name))
		}
		logger.Log(fmt.Sprintln(driver))
		logger.Log(fmt.Sprintln())
		break
	}
	parentWG.Done()
}

func buildGraph(driver string, users []User, connections map[string]map[string]int) *WeightedGraph {
	graph := NewGraph()
	nodes := graph.AddNodes(buildNodes(driver, users)...)

	// fmt.Printf("\nNodes: ")
	// fmt.Println(nodes)
	// fmt.Println("Edges")

	rwMutex.RLock()
	for src, connection := range connections {
		for des, distance := range connection {

			// TODO: if multiple drivers causes issue, check this condition
			if nodes[src] != nil && nodes[des] != nil && distance >= 0 {
				graph.AddEdge(nodes[src], nodes[des], int(distance)) // TODO: update distance to float32
			}
		}
	}
	rwMutex.RUnlock()

	return graph
}

func buildNodes(driver string, users []User) (nodeNames []string) {
	rwMutex.RLock()
	defer rwMutex.RUnlock()
	for _, user := range users {
		// Avoid other drivers while building graph
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
			rwMutex.Lock()
			globalData.activeUsers = append(globalData.activeUsers, u)
			rwMutex.Unlock()

		case cnr = <-nodesCh:
			rwMutex.Lock()
			globalData.connectedNodes[cnr.node] = cnr.value
			rwMutex.Unlock()

		}

	}
}

func calculateDistanceToExistingNodes(newUser User, existingUsers []User, nodesCh chan ConnectedNodesRequest) {
	var localWG sync.WaitGroup
	localData := make(map[string]int)
	var mu sync.Mutex

	for _, user := range existingUsers {
		if user.name == newUser.name || user.userType == Driver {
			continue
		}

		localWG.Add(1)
		go func(user User) {
			distance := getDistance(newUser.location, user.location)
			// Using mutex to avoid concurrent writes to the map
			mu.Lock()
			localData[user.name] = distance
			mu.Unlock()
			localWG.Done()
		}(user)
	}
	localWG.Wait()

	logger.Log(fmt.Sprintf("Distance matrix for user %s: %v\n", newUser.name, localData))
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
				location: generateRandomLocation(30),
				userType: Passenger,
			}
			localWG.Done()
		}(ActivePassengerCount)
	}
	localWG.Wait()
	// quitCh <- 1
	wg.Done()
}

func getDistance(src Location, dest Location) int {
	srcCoordinates := strconv.FormatFloat(src.Long, 'f', -1, 64) + "," + strconv.FormatFloat(src.Lat, 'f', -1, 64)
	destCoordinates := strconv.FormatFloat(dest.Long, 'f', -1, 64) + "," + strconv.FormatFloat(dest.Lat, 'f', -1, 64)

	url := "http://localhost:5000/route/v1/driving/" + srcCoordinates + ";" + destCoordinates + "?overview=false&alternatives=true&steps=false"
	resp, err := http.Get(url)
	if err != nil {
		logger.Log(fmt.Sprintf("Unable to get response from distance API. URL: %s\n", url))
		return -1
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var result DistanceResponse
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Log(fmt.Sprintln("Unable to unmarshal JSON"))
	}

	if len(result.Routes) > 0 {
		return int(result.Routes[0].Distance)
	}
	return -1
}

// https://gis.stackexchange.com/questions/25877/generating-random-locations-nearby
func generateRandomLocation(radiusInKm int) Location {
	x0, y0 := 50.565100, 9.686800                                  // lat long of HS Fulda
	radius := float64(float64(radiusInKm*1000) / float64(1113000)) // convert km into degress
	u, v := rand.Float64(), rand.Float64()
	w := radius * math.Sqrt(u)
	t := 2 * math.Pi * v
	x := w * math.Cos(t)
	y := w * math.Sin(t)
	x = x / math.Cos(y0)
	return Location{x + x0, y + y0}
}

func addHSFuldaNode(ch chan<- User) {
	ch <- User{
		name:     "HS",
		location: Location{50.565100, 9.686800},
		userType: Destination,
	}
}
