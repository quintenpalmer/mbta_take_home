package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func main() {
	api := ConcreteMBTAWebServer{}

	if err := print_light_and_heavy_rail_routes(api); err != nil {
		panic(err)
	}

	if err := print_stop_data(api); err != nil {
		panic(err)
	}

	if err := prompt_for_stops_to_route(api); err != nil {
		panic(err)
	}
}

type MBTAWebServer interface {
	GetRoutes(RouteRailType, RouteRailType) (RouteWrapper, error)
	GetStops(Route) (StopWrapper, error)
}

type ConcreteMBTAWebServer struct{}

func (c ConcreteMBTAWebServer) GetRoutes(type1 RouteRailType, type2 RouteRailType) (RouteWrapper, error) {
	url := fmt.Sprintf("https://api-v3.mbta.com/routes?filter[type]=%d,%d", type1, type2)
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

func (c ConcreteMBTAWebServer) GetStops(route Route) (StopWrapper, error) {
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
	ID        string        `json:"id"`
	Attribute StopAttribute `json:"attributes"`
}

type StopAttribute struct {
	Name string `json:"name"`
}

type RouteRailType int

// Per https://api-v3.mbta.com/docs/swagger/index.html#/Route/ApiWeb_RouteController_index
// The "Light Rail" is 0 and the "Heavy Rail" is 1.
// There are more rail types, but we are not concerning ourselves with it for this problem.
const (
	RouteRailTypeLightRail RouteRailType = iota
	RouteRailTypeHeavyRail
)

func print_light_and_heavy_rail_routes(api MBTAWebServer) error {
	names, err := list_light_and_heavy_rail_routes(api)
	if err != nil {
		return err
	}

	fmt.Println("The Heavy Rail and Light Rail Routes are:")
	for _, name := range names {
		fmt.Println(name)
	}
	fmt.Println("")

	return nil
}

func list_light_and_heavy_rail_routes(api MBTAWebServer) ([]string, error) {
	wrapper, err := get_heavy_and_light_routes(api)
	if err != nil {
		return nil, err
	}

	names := []string{}

	for _, route := range wrapper.Data {
		names = append(names, route.Attribute.LongName)
	}

	return names, nil
}

func get_heavy_and_light_routes(api MBTAWebServer) (RouteWrapper, error) {
	// Question 1 mentions how we can filter for the rail types on the query, or could filter after having
	// consumed the response. To save on retrieving data that we don't need, we are asking the server to filter
	// for us. Given how the filter types are documented with their own type, I feel this speaks reasonably well
	// as to what is happening with the query params going into the request.
	return api.GetRoutes(RouteRailTypeLightRail, RouteRailTypeHeavyRail)
}

type MinMaxData struct {
	Min      int
	MinRoute string
	Max      int
	MaxRoute string
}

func collect_stop_data(api MBTAWebServer) (MinMaxData, map[Stop][]Route, error) {
	wrapper, err := get_heavy_and_light_routes(api)
	if err != nil {
		return MinMaxData{}, nil, err
	}

	routeStops := map[Route][]Stop{}

	stopRoutes := map[Stop][]Route{}

	for _, route := range wrapper.Data {
		stops, err := api.GetStops(route)
		if err != nil {
			return MinMaxData{}, nil, err
		}
		routeStops[route] = stops.Data
		for _, stop := range stops.Data {
			stopRoutes[stop] = append(stopRoutes[stop], route)
		}
	}

	min := 999999
	minRoute := ""
	max := 0
	maxRoute := ""

	for route, stops := range routeStops {
		if len(stops) > max {
			maxRoute = route.Attribute.LongName
			max = len(stops)
		}
		if len(stops) < min {
			minRoute = route.Attribute.LongName
			min = len(stops)
		}
	}

	return MinMaxData{
		Min:      min,
		MinRoute: minRoute,
		Max:      max,
		MaxRoute: maxRoute,
	}, stopRoutes, nil
}

func print_stop_data(api MBTAWebServer) error {
	minMaxData, stopRoutes, err := collect_stop_data(api)
	if err != nil {
		return err
	}

	fmt.Println("Route with the minimum number of stops:")
	fmt.Printf("%s (with %d stops)\n", minMaxData.MinRoute, minMaxData.Min)
	fmt.Println("Route with the maximum number of stops:")
	fmt.Printf("%s (with %d stops)\n", minMaxData.MaxRoute, minMaxData.Max)
	fmt.Println("")

	fmt.Println("The following stops connect multiple routes:")
	for stop, routes := range stopRoutes {
		if len(routes) > 1 {
			fmt.Printf("Stop %s connects routes: %s\n", stop.Attribute.Name, build_route_list_name(routes))
		}
	}
	fmt.Println("")

	return nil
}

func build_route_list_name(routes []Route) string {
	list_name := ""
	first := true
	for _, route := range routes {
		if first {
			list_name = route.Attribute.LongName
			first = false
		} else {
			list_name = fmt.Sprintf("%s, %s", list_name, route.Attribute.LongName)
		}
	}
	return list_name
}

func prompt_for_stops_to_route(api MBTAWebServer) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Enter Starting Stop")
	startStop, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	fmt.Println("Enter Ending Stop")
	endStop, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	startStopName := strings.TrimSpace(startStop)
	endStopName := strings.TrimSpace(endStop)

	routes, err := routes_for_stop_to_stop(api, startStopName, endStopName)
	if err != nil {
		return err
	}

	if len(routes) > 0 {
		fmt.Printf("Take the following routes to get from %s to %s:\n", startStopName, endStopName)
		for _, route := range routes {
			fmt.Println(route.Attribute.LongName)
		}
	} else {
		fmt.Printf("The path from %s to %s is to take no routes, as they are the same path.\n", startStopName, endStopName)
	}

	return nil
}

var (
	ErrNoStartStop = errors.New("could not find start stop")
	ErrNoEndStop   = errors.New("could not find end stop")
)

func routes_for_stop_to_stop(api MBTAWebServer, startStopName string, endStopName string) ([]Route, error) {
	wrapper, err := get_heavy_and_light_routes(api)
	if err != nil {
		return nil, err
	}

	routeStops := map[Route][]Stop{}
	stopRoutes := map[Stop][]Route{}

	startStop := Stop{}
	endStop := Stop{}

	for _, route := range wrapper.Data {
		stops, err := api.GetStops(route)
		if err != nil {
			return nil, err
		}
		routeStops[route] = stops.Data
		for _, stop := range stops.Data {
			stopRoutes[stop] = append(stopRoutes[stop], route)
			if stop.Attribute.Name == startStopName {
				startStop = stop
			}
			if stop.Attribute.Name == endStopName {
				endStop = stop
			}
		}
	}

	if startStop.ID == "" {
		return nil, ErrNoStartStop
	}
	if endStop.ID == "" {
		return nil, ErrNoEndStop
	}

	routes, err := explore_routes_and_stops(routeStops, stopRoutes, startStop, endStop, []Route{}, map[Route]struct{}{})
	if err != nil {
		return nil, err
	}

	return routes, nil
}

var ErrNoPath = errors.New("no path to end stop from this branch")

func explore_routes_and_stops(routeStops map[Route][]Stop, stopRoutes map[Stop][]Route, currentStop Stop, endStop Stop, currentRoutes []Route, explored map[Route]struct{}) ([]Route, error) {
	if currentStop == endStop {
		return currentRoutes, nil
	}

	for _, route := range stopRoutes[currentStop] {
		for _, subStop := range routeStops[route] {
			if subStop == endStop {
				return append(currentRoutes, route), nil
			}
		}
		for _, subStop := range routeStops[route] {
			if _, ok := explored[route]; !ok {
				subExplored := map[Route]struct{}{}
				for exploredRoute := range explored {
					subExplored[exploredRoute] = struct{}{}
				}
				subExplored[route] = struct{}{}
				ret, err := explore_routes_and_stops(routeStops, stopRoutes, subStop, endStop, append(currentRoutes, route), subExplored)
				if err == nil {
					return ret, nil
				}
			}
		}
	}

	return []Route{}, ErrNoPath
}
