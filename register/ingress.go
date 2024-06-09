package register

import (
	guidehttp "github.com/advanced-go/guidance/http"
	guidemod "github.com/advanced-go/guidance/module"
	observhttp "github.com/advanced-go/observation/http"
	observmod "github.com/advanced-go/observation/module"
	searchhttp "github.com/advanced-go/search/http"
	searchmod "github.com/advanced-go/search/module"
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
