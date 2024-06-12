package register

import (
	"errors"
	"fmt"
	"github.com/advanced-go/guidance/resiliency1"
	"github.com/advanced-go/search/google"
	"github.com/advanced-go/search/yahoo"
	"github.com/advanced-go/stdlib/controller"
	"github.com/advanced-go/stdlib/controller2"
	"github.com/advanced-go/stdlib/core"
)

func EgressController() error {
	// Search Google and Yahoo package's egress
	status := controller2.RegisterControllerFromConfig(google.EgressRoute(), nil)
	if !status.OK() {
		return status.Err
	}
	status = controller2.RegisterControllerFromConfig(yahoo.EgressRoute(), nil)
	if !status.OK() {
		return status.Err
	}
	// Guidance resiliency1 package's egress
	status = controller2.RegisterControllerFromConfig(resiliency1.EgressRoute(), nil)
	if !status.OK() {
		return status.Err
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
