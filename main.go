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

type User struct {
	name     string
	location Location
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
	// assignDriver() to take activeusers at that time, create copy and
	wg.Add(1)
	go addDriver(globalData.activeUsersCh)
	wg.Wait()
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
			}
			localWG.Done()
		}(ActivePassengerCount)
	}
	localWG.Wait()
	// quitCh <- 1
	wg.Done()
}

func addDriver(ch chan<- User) {
	ActiveDriverCount += 1
	ch <- User{
		name:     "d" + strconv.Itoa(ActiveDriverCount),
		location: generateRandomLocation(),
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
	}
}
