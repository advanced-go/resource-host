package register

import (
	"errors"
	"fmt"
	dochttp "github.com/advanced-go/documents/http"
	"github.com/advanced-go/guidance/resiliency1"
	"github.com/advanced-go/observation/timeseries1"
	"github.com/advanced-go/search/google"
	"github.com/advanced-go/search/yahoo"
	"github.com/advanced-go/stdlib/controller"
	"github.com/advanced-go/stdlib/core"
	timehttp "github.com/advanced-go/timeseries/http"
)

func EgressExchange() error {
	// Search package's egress
	err := register(google.RouteName, google.EgressRoute, nil)
	if err != nil {
		return err
	}
	err = register(yahoo.RouteName, yahoo.EgressRoute, nil)
	if err != nil {
		return err
	}
	// Guidance package's egress
	err = register(resiliency1.RouteName, resiliency1.EgressRoute, dochttp.Exchange)
	if err != nil {
		return err
	}
	// Observation package's egress
	err = register(timeseries1.RouteName, timeseries1.EgressRoute, timehttp.Exchange)
	if err != nil {
		return err
	}
	return nil
}

func routeNameError(routeName string) error {
	return errors.New(fmt.Sprintf("error: route name is invalid: %v", routeName))
}

func register(routeName string, fn func(string) (*controller.Config, bool), ex core.HttpExchange) error {
	cfg, ok := fn(routeName)
	if !ok {
		return routeNameError(routeName)
	}
	status := controller.RegisterControllerFromConfig(cfg, ex)
	if !status.OK() {
		return status.Err
	}
	return nil
}
