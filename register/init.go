package register

import (
	"errors"
	"fmt"
	dochttp "github.com/advanced-go/documents/http"
	guidehttp "github.com/advanced-go/guidance/http"
	guidemod "github.com/advanced-go/guidance/module"
	"github.com/advanced-go/guidance/resiliency1"
	observhttp "github.com/advanced-go/observation/http"
	observmod "github.com/advanced-go/observation/module"
	"github.com/advanced-go/observation/timeseries1"
	"github.com/advanced-go/search/google"
	searchhttp "github.com/advanced-go/search/http"
	searchmod "github.com/advanced-go/search/module"
	"github.com/advanced-go/search/yahoo"
	"github.com/advanced-go/stdlib/controller"
	"github.com/advanced-go/stdlib/core"
	"github.com/advanced-go/stdlib/host"
	timehttp "github.com/advanced-go/timeseries/http"
)

func IngressExchanges() error {
	err := host.RegisterExchange(searchmod.Authority, host.NewAccessLogIntermediary(searchmod.RouteName, searchhttp.Exchange))
	if err != nil {
		return err
	}
	err = host.RegisterExchange(guidemod.Authority, host.NewAccessLogIntermediary(guidemod.RouteName, guidehttp.Exchange))
	if err != nil {
		return err
	}
	err = host.RegisterExchange(observmod.Authority, host.NewAccessLogIntermediary(observmod.RouteName, observhttp.Exchange))
	if err != nil {
		return err
	}
	return nil
}

func EgressControllers() error {
	// Search package's egress
	err := register(google.RouteName, google.Route, nil)
	if err != nil {
		return err
	}
	err = register(yahoo.RouteName, yahoo.Route, nil)
	if err != nil {
		return err
	}
	// Guidance package's egress
	err = register(resiliency1.RouteName, resiliency1.Route, dochttp.Exchange)
	if err != nil {
		return err
	}
	// Observation package's egress
	err = register(timeseries1.RouteName, timeseries1.Route, timehttp.Exchange)
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
