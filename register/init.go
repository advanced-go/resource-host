package register

import (
	dochttp "github.com/advanced-go/documents/http"
	guidehttp "github.com/advanced-go/guidance/http"
	guidemod "github.com/advanced-go/guidance/module"
	"github.com/advanced-go/guidance/resiliency1"
	observhttp "github.com/advanced-go/observation/http"
	observmod "github.com/advanced-go/observation/module"
	"github.com/advanced-go/search/google"
	searchhttp "github.com/advanced-go/search/http"
	searchmod "github.com/advanced-go/search/module"
	"github.com/advanced-go/search/yahoo"
	"github.com/advanced-go/stdlib/controller"
	"github.com/advanced-go/stdlib/host"
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
	status := controller.RegisterControllerFromConfig(google.Route, nil)
	if !status.OK() {
		return status.Err
	}
	status = controller.RegisterControllerFromConfig(yahoo.Route, nil)
	if !status.OK() {
		return status.Err
	}
	status = controller.RegisterControllerFromConfig(resiliency1.Route, dochttp.Exchange)
	if !status.OK() {
		return status.Err
	}
	return nil
}
