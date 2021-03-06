package item

import (
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/labels"
	kv1 "k8s.io/client-go/listers/core/v1"

	"github.com/amadeusitgroup/kubervisor/pkg/pod"

	activator "github.com/amadeusitgroup/kubervisor/pkg/activate"
	apiv1 "github.com/amadeusitgroup/kubervisor/pkg/api/kubervisor/v1"
	"github.com/amadeusitgroup/kubervisor/pkg/breaker"
)

// New return new KubervisorServiceItem instance
func New(bc *apiv1.KubervisorService, cfg *Config) (Interface, error) {
	if cfg.customFactory != nil {
		return cfg.customFactory(bc, cfg)
	}

	activateConfig := activator.FactoryConfig{
		Config: activator.Config{
			ActivatorStrategyConfig: bc.Spec.Activator,
			Selector:                cfg.Selector,
			BreakerName:             bc.Name,
			PodControl:              cfg.PodControl,
			PodLister:               cfg.PodLister.Pods(bc.Namespace),
			Logger:                  cfg.Logger,
		},
	}
	activatorInterface, err := activator.New(activateConfig)
	if err != nil {
		return nil, err
	}

	breakerConfig := breaker.FactoryConfig{
		Config: breaker.Config{
			KubervisorServiceName: bc.Name,
			BreakerStrategyConfig: bc.Spec.Breaker,
			Selector:              cfg.Selector,
			PodControl:            cfg.PodControl,
			PodLister:             cfg.PodLister.Pods(bc.Namespace),
			Logger:                cfg.Logger,
			BreakerName:           bc.Name,
		},
	}
	breakerInterface, err := breaker.New(breakerConfig)
	if err != nil {
		return nil, err
	}
	return &KubervisorServiceItem{
		name:      bc.Name,
		namespace: bc.Namespace,
		activator: activatorInterface,
		breaker:   breakerInterface,
	}, nil

}

// Config Item factory configuration
type Config struct {
	Selector   labels.Selector
	PodLister  kv1.PodLister
	PodControl pod.ControlInterface
	Logger     *zap.Logger

	customFactory Factory
}

//Factory functor for Interface
type Factory func(bc *apiv1.KubervisorService, cfg *Config) (Interface, error)

var _ Factory = New
