package main

import (
	"errors"
	"reflect"
	"testing"
)

type MockMBTAWebServer struct {
	RecvType1               *RouteRailType
	RecvType2               *RouteRailType
	ReturnRouteWrapper      RouteWrapper
	ReturnRouteWrapperError error

	RecvRoutes             []Route
	ReturnStopWrapper      map[string]StopWrapper
	ReturnStopWrapperError error
}

func (c *MockMBTAWebServer) GetRoutes(type1 RouteRailType, type2 RouteRailType) (RouteWrapper, error) {
	c.RecvType1 = &type1
	c.RecvType2 = &type2
	return c.ReturnRouteWrapper, c.ReturnRouteWrapperError
}

func (c *MockMBTAWebServer) GetStops(route Route) (StopWrapper, error) {
	c.RecvRoutes = append(c.RecvRoutes, route)
	return c.ReturnStopWrapper[route.ID], c.ReturnStopWrapperError
}

func Test_list_light_and_heavy_rail_routes(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		mockAPI := &MockMBTAWebServer{
			ReturnRouteWrapper: RouteWrapper{
				Data: []Route{
					{
						Attribute: RouteAttribute{
							LongName: "mock route name 1",
						},
					},
					{
						Attribute: RouteAttribute{
							LongName: "mock route name 2",
						},
					},
				},
			},
		}

		expected := []string{"mock route name 1", "mock route name 2"}

		names, err := list_light_and_heavy_rail_routes(mockAPI)
		if err != nil {
			t.Error("did not expect an error")
		}
		if !reflect.DeepEqual(expected, names) {
			t.Errorf("expected %s to be equal to %s", expected, names)
		}
	})

	t.Run("sad path", func(t *testing.T) {
		myErr := errors.New("custom mock error")

		mockAPI := &MockMBTAWebServer{
			ReturnRouteWrapperError: myErr,
		}

		_, err := list_light_and_heavy_rail_routes(mockAPI)
		if err != myErr {
			t.Errorf("expected error %s to be %s", myErr, err)
		}
	})
}

func Test_get_heavy_and_light_routes(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		wrapper := RouteWrapper{
			Data: []Route{
				{
					Attribute: RouteAttribute{
						LongName: "mock route name 1",
					},
				},
				{
					Attribute: RouteAttribute{
						LongName: "mock route name 2",
					},
				},
			},
		}

		mockAPI := &MockMBTAWebServer{
			ReturnRouteWrapper: wrapper,
		}

		routes, err := get_heavy_and_light_routes(mockAPI)
		if err != nil {
			t.Error("did not expect an error")
		}
		if !reflect.DeepEqual(wrapper, routes) {
			t.Errorf("expected %s to be equal to %s", wrapper, routes)
		}
		if *mockAPI.RecvType1 != RouteRailTypeLightRail {
			t.Errorf("expected received %d to be %d", mockAPI.RecvType1, RouteRailTypeLightRail)
		}
		if *mockAPI.RecvType2 != RouteRailTypeHeavyRail {
			t.Errorf("expected received %d to be %d", mockAPI.RecvType2, RouteRailTypeHeavyRail)
		}
	})

	t.Run("sad path", func(t *testing.T) {
		myErr := errors.New("custom mock error")

		mockAPI := &MockMBTAWebServer{
			ReturnRouteWrapperError: myErr,
		}

		_, err := get_heavy_and_light_routes(mockAPI)
		if err != myErr {
			t.Errorf("expected error %s to be %s", myErr, err)
		}
	})
}

func Test_collect_stop_data(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		mockAPI := &MockMBTAWebServer{
			ReturnRouteWrapper: RouteWrapper{
				Data: []Route{
					{
						ID: "route key 1",
						Attribute: RouteAttribute{
							LongName: "mock route name 1",
						},
					},
					{
						ID: "route key 2",
						Attribute: RouteAttribute{
							LongName: "mock route name 2",
						},
					},
				},
			},
			ReturnStopWrapper: map[string]StopWrapper{
				"route key 1": StopWrapper{
					Data: []Stop{
						{
							ID: "stop key 1",
							Attribute: StopAttribute{
								Name: "mock stop name 1",
							},
						},
					},
				},
				"route key 2": StopWrapper{
					Data: []Stop{
						{
							ID: "stop key 1",
							Attribute: StopAttribute{
								Name: "mock stop name 1",
							},
						},
						{
							ID: "stop key 2",
							Attribute: StopAttribute{
								Name: "mock stop name 2",
							},
						},
					},
				},
			},
		}

		expectedMinMaxData := MinMaxData{
			Min:      1,
			MinRoute: "mock route name 1",
			Max:      2,
			MaxRoute: "mock route name 2",
		}

		expectedStopRoutes := map[Stop][]Route{
			{
				ID: "stop key 1",
				Attribute: StopAttribute{
					Name: "mock stop name 1",
				},
			}: []Route{
				{
					ID: "route key 1",
					Attribute: RouteAttribute{
						LongName: "mock route name 1",
					},
				},
				{
					ID: "route key 2",
					Attribute: RouteAttribute{
						LongName: "mock route name 2",
					},
				},
			},
			{
				ID: "stop key 2",
				Attribute: StopAttribute{
					Name: "mock stop name 2",
				},
			}: []Route{
				{
					ID: "route key 2",
					Attribute: RouteAttribute{
						LongName: "mock route name 2",
					},
				},
			},
		}

		minMaxData, stopRoutes, err := collect_stop_data(mockAPI)
		if err != nil {
			t.Error("did not expect an error")
		}
		if expectedMinMaxData != minMaxData {
			t.Errorf("expected %+v to be equal to %+v", expectedMinMaxData, minMaxData)
		}
		if !reflect.DeepEqual(expectedStopRoutes, stopRoutes) {
			t.Errorf("expected %+v to be equal to %+v", expectedStopRoutes, stopRoutes)
		}
	})

	t.Run("sad path - route lookup fails", func(t *testing.T) {
		myErr := errors.New("custom mock error")

		mockAPI := &MockMBTAWebServer{
			ReturnRouteWrapperError: myErr,
		}

		_, _, err := collect_stop_data(mockAPI)
		if err != myErr {
			t.Errorf("expected error %s to be %s", myErr, err)
		}
	})

	t.Run("sad path - stop lookup fails", func(t *testing.T) {
		myErr := errors.New("custom mock error")

		mockAPI := &MockMBTAWebServer{
			ReturnRouteWrapper: RouteWrapper{
				Data: []Route{
					{
						Attribute: RouteAttribute{
							LongName: "mock route name 1",
						},
					},
					{
						Attribute: RouteAttribute{
							LongName: "mock route name 2",
						},
					},
				},
			},
			ReturnStopWrapperError: myErr,
		}

		_, _, err := collect_stop_data(mockAPI)
		if err != myErr {
			t.Errorf("expected error %s to be %s", myErr, err)
		}
	})
}

func Test_build_route_list_name(t *testing.T) {
	t.Run("happy path - no values", func(t *testing.T) {
		input := []Route{}

		expected := ""

		result := build_route_list_name(input)
		if expected != result {
			t.Errorf("expected %v to be equal to %v", expected, result)
		}
	})

	t.Run("happy path - one value", func(t *testing.T) {
		input := []Route{{Attribute: RouteAttribute{LongName: "route name 1"}}}

		expected := "route name 1"

		result := build_route_list_name(input)
		if expected != result {
			t.Errorf("expected %v to be equal to %v", expected, result)
		}
	})

	t.Run("happy path - two values", func(t *testing.T) {
		input := []Route{{Attribute: RouteAttribute{LongName: "route name 1"}}, {Attribute: RouteAttribute{LongName: "route name 2"}}}

		expected := "route name 1, route name 2"

		result := build_route_list_name(input)
		if expected != result {
			t.Errorf("expected %v to be equal to %v", expected, result)
		}
	})
}

func Test_explore_routes_and_stops(t *testing.T) {
	t.Run("happy path - same start and end", func(t *testing.T) {
		routeStops := map[Route][]Stop{}
		stopRoutes := map[Stop][]Route{}
		currentStop := Stop{ID: "stop id 1"}
		endStop := Stop{ID: "stop id 1"}
		currentRoutes := []Route{}
		explored := map[Route]struct{}{}

		expected := []Route{}

		found, err := explore_routes_and_stops(routeStops, stopRoutes, currentStop, endStop, currentRoutes, explored)
		if err != nil {
			t.Error("did not expect an error")
		}
		if !reflect.DeepEqual(expected, found) {
			t.Errorf("expected %s to be equal to %s", expected, found)
		}
	})

	t.Run("happy path - on same route", func(t *testing.T) {
		routeStops := map[Route][]Stop{
			Route{ID: "route id 1"}: []Stop{
				{ID: "stop id 1"},
				{ID: "stop id 2"},
			},
		}
		stopRoutes := map[Stop][]Route{
			Stop{ID: "stop id 1"}: []Route{
				{ID: "route id 1"},
			},
			Stop{ID: "stop id 2"}: []Route{
				{ID: "route id 1"},
			},
		}
		currentStop := Stop{ID: "stop id 1"}
		endStop := Stop{ID: "stop id 2"}
		currentRoutes := []Route{}
		explored := map[Route]struct{}{}

		expected := []Route{{ID: "route id 1"}}

		found, err := explore_routes_and_stops(routeStops, stopRoutes, currentStop, endStop, currentRoutes, explored)
		if err != nil {
			t.Error("did not expect an error")
		}
		if !reflect.DeepEqual(expected, found) {
			t.Errorf("expected %s to be equal to %s", expected, found)
		}
	})

	t.Run("happy path - one route hop", func(t *testing.T) {
		routeStops := map[Route][]Stop{
			Route{ID: "route id 1"}: []Stop{
				{ID: "stop id 1"},
				{ID: "stop id 2"},
			},
			Route{ID: "route id 2"}: []Stop{
				{ID: "stop id 2"},
				{ID: "stop id 3"},
			},
		}
		stopRoutes := map[Stop][]Route{
			Stop{ID: "stop id 1"}: []Route{
				{ID: "route id 1"},
			},
			Stop{ID: "stop id 2"}: []Route{
				{ID: "route id 1"},
				{ID: "route id 2"},
			},
			Stop{ID: "stop id 3"}: []Route{
				{ID: "route id 2"},
			},
		}
		currentStop := Stop{ID: "stop id 1"}
		endStop := Stop{ID: "stop id 3"}
		currentRoutes := []Route{}
		explored := map[Route]struct{}{}

		expected := []Route{{ID: "route id 1"}, {ID: "route id 2"}}

		found, err := explore_routes_and_stops(routeStops, stopRoutes, currentStop, endStop, currentRoutes, explored)
		if err != nil {
			t.Error("did not expect an error")
		}
		if !reflect.DeepEqual(expected, found) {
			t.Errorf("expected %s to be equal to %s", expected, found)
		}
	})

	t.Run("sad path - one route hop - missing stop 2 route 2 connection", func(t *testing.T) {
		routeStops := map[Route][]Stop{
			Route{ID: "route id 1"}: []Stop{
				{ID: "stop id 1"},
				{ID: "stop id 2"},
			},
			Route{ID: "route id 2"}: []Stop{
				{ID: "stop id 2"},
				{ID: "stop id 3"},
			},
		}
		stopRoutes := map[Stop][]Route{
			Stop{ID: "stop id 1"}: []Route{
				{ID: "route id 1"},
			},
			Stop{ID: "stop id 2"}: []Route{
				{ID: "route id 1"},
			},
			Stop{ID: "stop id 3"}: []Route{
				{ID: "route id 2"},
			},
		}
		currentStop := Stop{ID: "stop id 1"}
		endStop := Stop{ID: "stop id 3"}
		currentRoutes := []Route{}
		explored := map[Route]struct{}{}

		_, err := explore_routes_and_stops(routeStops, stopRoutes, currentStop, endStop, currentRoutes, explored)
		if err != ErrNoPath {
			t.Errorf("expected error %s to be %s", err, ErrNoPath)
		}
	})
}
