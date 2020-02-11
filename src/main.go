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

	if err := print_stop_data(); err != nil {
		panic(err)
	}
}

type RouteWrapper struct {
	Data []Route `json:"data"`
}

type Route struct {
	ID        string         `json:"id"`
	Attribute RouteAttribute `json:"attributes"`
}

type RouteAttribute struct {
	LongName string `json:"long_name"`
}

type StopWrapper struct {
	Data []Stop `json:"data"`
}

type Stop struct {
	ID string `json:"id"`
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

func print_stop_data() error {
	wrapper, err := get_heavy_and_light_routes()
	if err != nil {
		return err
	}

	routeStops := map[Route]StopWrapper{}

	for _, route := range wrapper.Data {
		stops, err := get_route_stops(route)
		if err != nil {
			return err
		}
		routeStops[route] = stops
	}

	min := 999999
	minRoute := ""
	max := 0
	maxRoute := ""

	for route, stops := range routeStops {
		if len(stops.Data) > max {
			maxRoute = route.Attribute.LongName
			max = len(stops.Data)
		}
		if len(stops.Data) < min {
			minRoute = route.Attribute.LongName
			min = len(stops.Data)
		}
	}

	fmt.Println("Route with the minimum number of stops:")
	fmt.Printf("%s (with %d stops)\n", minRoute, min)
	fmt.Println("Route with the maximum number of stops:")
	fmt.Printf("%s (with %d stops)\n", maxRoute, max)

	return nil
}

func get_route_stops(route Route) (StopWrapper, error) {
	// We want the count of stops for each route. From what I am reading on:
	// https://api-v3.mbta.com/docs/swagger/index.html#/Stop/ApiWeb_StopController_index
	// we can only display the route information if we give exactly one route to filter for.
	// This means that we have to request the stops for each route, but given that there are only
	// 8 light and heavy routes total, this shouldn't overload their servers or cause a time-out
	// for this client.
	url := fmt.Sprintf("https://api-v3.mbta.com/stops?filter[route]=%s", route.ID)
	resp, err := http.Get(url)
	if err != nil {
		return StopWrapper{}, err
	}
	wrapper := StopWrapper{}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&wrapper); err != nil {
		return StopWrapper{}, err
	}

	return wrapper, nil
}
