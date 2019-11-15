package weights

import (
	"github.com/pkg/errors"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
)

func StepWeights(virtualService *v1alpha3.VirtualService, stepWeight int32) error {
	primaryWeight, canaryWeight, err := GetWeights(*virtualService)
	if err != nil {
		return errors.Wrapf(err, "failed to get virtual service weights for virtualService %v", virtualService.Name)
	}

	// once primary weight is below 0, we just monitor all the traffic going to canary
	if primaryWeight > 0 {
		if err := SetWeights(virtualService, primaryWeight-stepWeight, canaryWeight+stepWeight); err != nil {
			return errors.Wrapf(err, "failed to set weights for virtual service %v", virtualService.Name)
		}
	}

	return nil
}

// gets the weights for the canary and primary destinations
// only checks the first per-port route, assuming all routes have the same weights
func GetWeights(virtualService v1alpha3.VirtualService) (int32, int32, error) {
	// for shift weights on each per-port route
	for _, httpPort := range virtualService.Spec.Http {
		if len(httpPort.Route) != 2 {
			return 0, 0, errors.Errorf("expected 2 routes on http rule %v, found %v", httpPort.Name, len(virtualService.Spec.Http))
		}
		primaryRoute, canaryRoute := httpPort.Route[0], httpPort.Route[1]

		// set primary
		return primaryRoute.Weight, canaryRoute.Weight, nil
	}

	return 0, 0, errors.Errorf("no routes found")
}

// sets the weights to the canary and primary destinations
func SetWeights(virtualService *v1alpha3.VirtualService, primaryWeight, canaryWeight int32) error {
	// for shift weights on each per-port route
	for _, httpPort := range virtualService.Spec.Http {
		if len(httpPort.Route) != 2 {
			return errors.Errorf("expected 2 routes on http rule %v, found %v", httpPort.Name, len(virtualService.Spec.Http))
		}
		primaryRoute, canaryRoute := httpPort.Route[0], httpPort.Route[1]

		// set primary
		if primaryWeight < 0 {
			primaryWeight = 0
		}
		primaryRoute.Weight = primaryWeight
		// set canary
		if canaryWeight > 100 {
			canaryWeight = 100
		}
		canaryRoute.Weight = canaryWeight
	}

	return nil
}
