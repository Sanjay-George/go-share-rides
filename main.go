package main

import (
	"fmt"
	"math/rand"
)

const (
	HSFuldaUser = "HS"
)

var HSFuldaCoordinates = Location{50.565100, 9.686800}

type GlobalMap struct {
	// mu   sync.Mutex
	data    map[string]map[string]float32
	readCh  chan map[string]float32
	writeCh chan map[string]float32
}

type Location struct {
	Lat, Long float64
}

func main() {
	globalMap := GlobalMap{readCh: make(chan map[string]float32, 1000), writeCh: make(chan map[string]float32)}

	fmt.Println("Hello! Let's build something awesome!")

	go addPassenger("p1", generateRandomLocation(), globalMap.writeCh)
	go addPassenger("p2", generateRandomLocation(), globalMap.writeCh)

	handleGlobalMapReadWrites(&globalMap)
}

func handleGlobalMapReadWrites(globalMap *GlobalMap) {
	var localMap = make(map[string]float32)
	for {
		select {
		case localMap = <-globalMap.writeCh:

		}
	}
}

func getDistance(src Location, dest Location) float32 {
	// TODO: make API call to get shortest distance by lat, long
	return float32(rand.Intn(100))
}

// TODO:
func generateRandomLocation() Location {
	return Location{0, 0}
}

func addPassenger(name string, location Location, ch chan map[string]float32) {
	x := getDistance(location, HSFuldaCoordinates)

	localMap := make(map[string]float64)

	// TODO: create localMap, add all values to the map, return to main method. Main method will add it to GlobalMap

	// TODO: calculate distance to each user and update the globalMap

}

// func addDriver()
