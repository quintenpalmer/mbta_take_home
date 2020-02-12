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
	ReturnStopWrapper      StopWrapper
	ReturnStopWrapperError error
}

func (c *MockMBTAWebServer) GetRoutes(type1 RouteRailType, type2 RouteRailType) (RouteWrapper, error) {
	c.RecvType1 = &type1
	c.RecvType2 = &type2
	return c.ReturnRouteWrapper, c.ReturnRouteWrapperError
}

func (c *MockMBTAWebServer) GetStops(route Route) (StopWrapper, error) {
	c.RecvRoutes = append(c.RecvRoutes, route)
	return c.ReturnStopWrapper, c.ReturnStopWrapperError
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
