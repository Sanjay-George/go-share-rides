package main

type Location struct {
	Lat, Long float64
}
type ConnectedNodesRequest struct {
	node  string
	value map[string]uint32
}

const (
	Destination = iota
	Driver      = iota
	Passenger   = iota
)

type User struct {
	name     string
	location Location
	userType uint8
}

type GlobalState struct {
	activeUsers      []User
	activeUsersCh    chan User
	connectedNodes   map[string]map[string]uint32
	connectedNodesCh chan ConnectedNodesRequest
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
