package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	if err := list_light_and_heavy_rail_routes(); err != nil {
		panic(err)
	}
}

type RouteWrapper struct {
	Data []Route `json:"data"`
}

type Route struct {
	Attribute Attribute `json:"attributes"`
}

type Attribute struct {
	LongName string `json:"long_name"`
}

type RouteRailType int

// Per https://api-v3.mbta.com/docs/swagger/index.html#/Route/ApiWeb_RouteController_index
// The "Light Rail" is 0 and the "Heavy Rail" is 1.
// There are more rail types, but we are not concerning ourselves with it for this problem.
const (
	RouteRailTypeLightRail RouteRailType = iota
	RouteRailTypeHeavyRail
)

func list_light_and_heavy_rail_routes() error {
	wrapper, err := get_heavy_and_light_routes()
	if err != nil {
		return err
	}

	fmt.Println("The Heavy Rail and Light Rail Routes are:")
	for _, route := range wrapper.Data {
		fmt.Println(route.Attribute.LongName)
	}
	fmt.Println("")

	return nil
}

func get_heavy_and_light_routes() (RouteWrapper, error) {
	// Question 1 mentions how we can filter for the rail types on the query, or could filter after having
	// consumed the response. To save on retrieving data that we don't need, we are asking the server to filter
	// for us. Given how the filter types are documented with their own type, I feel this speaks reasonably well
	// as to what is happening with the query params going into the request.
	url := fmt.Sprintf("https://api-v3.mbta.com/routes?filter[type]=%d,%d", RouteRailTypeLightRail, RouteRailTypeHeavyRail)
	resp, err := http.Get(url)
	if err != nil {
		return RouteWrapper{}, err
	}
	wrapper := RouteWrapper{}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&wrapper); err != nil {
		return RouteWrapper{}, err
	}

	return wrapper, nil
}
