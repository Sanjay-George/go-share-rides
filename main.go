package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
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
)

type Location struct {
	Lat, Long float64
}
type ConnectedNodesRequest struct {
	node  string
	value map[string]int
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
	connectedNodes   map[string]map[string]int
	connectedNodesCh chan ConnectedNodesRequest
}

// TODO: Ensure global state is not directly modified by any other goroutine
// Use channels for communication and modify from main goroutine
var globalData = GlobalState{
	activeUsers:      make([]User, 0, 100), // capacity of 100 users initially
	connectedNodes:   make(map[string]map[string]int),
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

	rand.Seed(time.Now().UnixNano())

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

func assignPassengers(driver string, users []User, connections map[string]map[string]int) {
	graph := buildGraph(driver, users, connections)
	fmt.Printf("\nGraph")
	fmt.Println(graph)

	maxDistance := int(float32(connections[driver][HSFuldaUsername]) * MultiplicationFactor) // TODO: find optimal multiplication factor
	FindOptimalPath(graph, driver, HSFuldaUsername, int(maxDistance))

	for _, node := range graph.Nodes {
		if node.name == HSFuldaUsername {
			fmt.Printf("Optimal path from %s to %s covers %d m with emission %d units\n", driver, HSFuldaUsername, node.shortestDistance, node.GetEmissionValue())
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

func buildGraph(driver string, users []User, connections map[string]map[string]int) *WeightedGraph {
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
	localData := make(map[string]int)
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
				location: generateRandomLocation(5),
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
		location: generateRandomLocation(30),
		userType: Driver,
	}
	wg.Done()
}

type DistanceResponse struct {
	Code   string `json:"code"`
	Routes []struct {
		Legs []struct {
			Steps    []interface{} `json:"steps"`
			Summary  string        `json:"summary"`
			Weight   float64       `json:"weight"`
			Duration float64       `json:"duration"`
			Distance float64       `json:"distance"`
		} `json:"legs"`
		WeightName string  `json:"weight_name"`
		Weight     float64 `json:"weight"`
		Duration   float64 `json:"duration"`
		Distance   float64 `json:"distance"`
	} `json:"routes"`
	Waypoints []struct {
		Hint     string    `json:"hint"`
		Distance float64   `json:"distance"`
		Name     string    `json:"name"`
		Location []float64 `json:"location"`
	} `json:"waypoints"`
}

func getDistance(src Location, dest Location) int {
	// TODO: fetch data from OSRM
	// http://localhost:5000/route/v1/driving/9.685991642142039,50.5650744;9.6800597,50.5552363?overview=false&alternatives=true&steps=false

	time.Sleep(1000 * time.Millisecond)

	// distance := float32(rand.Intn(100))

	srcCoordinates := strconv.FormatFloat(src.Long, 'f', -1, 64) + "," + strconv.FormatFloat(src.Lat, 'f', -1, 64)
	destCoordinates := strconv.FormatFloat(dest.Long, 'f', -1, 64) + "," + strconv.FormatFloat(dest.Lat, 'f', -1, 64)

	url := "http://localhost:5000/route/v1/driving/" + srcCoordinates + ";" + destCoordinates + "?overview=false&alternatives=true&steps=false"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Unable to get response from distance API")
		return math.MaxInt
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(body))

	var result DistanceResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Unable to unmarshal JSON")
	}

	if len(result.Routes) > 0 {
		return int(result.Routes[0].Distance)
	}
	return math.MaxInt
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
