package kubernetes

import (
	"context"
	"fmt"
	"time"

	"github.com/MouseHatGames/mice/discovery"
	"github.com/MouseHatGames/mice/logger"
	"github.com/MouseHatGames/mice/options"
	"github.com/patrickmn/go-cache"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type endpoints = []string

type k8sDiscovery struct {
	log            logger.Logger
	opts           *k8sOptions
	endpointClient corev1.EndpointsInterface

	cache *cache.Cache
}

func Discovery(opts ...K8sOption) options.Option {
	return func(o *options.Options) {
		k8opt := &k8sOptions{
			Namespace: "default",
			CacheTime: 1 * time.Minute,
		}

		for _, opt := range opts {
			opt(k8opt)
		}

		if k8opt.Selection == nil {
			RoundRobin()(k8opt)
		}

		o.Discovery = &k8sDiscovery{
			opts:  k8opt,
			log:   o.Logger.GetLogger("k8s"),
			cache: cache.New(k8opt.CacheTime, 5*time.Minute),
		}
	}
}

func (d *k8sDiscovery) Start() error {
	d.log.Debugf("getting in-cluster config")

	config, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("get in-cluster config: %w", err)
	}

	d.log.Debugf("creating client")

	cl, err := corev1.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("get clientset: %w", err)
	}

	d.endpointClient = cl.Endpoints(d.opts.Namespace)

	return nil
}

func (d *k8sDiscovery) getEndpoints(name string) ([]string, error) {
	if val, ok := d.cache.Get(name); ok {
		return val.([]string), nil
	}

	d.log.Debugf("fetching %s endpoints", name)
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	endpoints, err := d.endpointClient.List(ctx, v1.ListOptions{LabelSelector: fmt.Sprintf("mice=%s", name)})
	if err != nil {
		return nil, fmt.Errorf("list pods: %w", err)
	}

	hosts := make([]string, 0, len(endpoints.Items))

	for _, ep := range endpoints.Items {
		for _, subset := range ep.Subsets {
			for _, addr := range subset.Addresses {
				hosts = append(hosts, addr.IP)
			}
		}
	}

	d.log.Debugf("got %d endpoints for %s in %s", len(hosts), name, time.Now().Sub(start))

	d.cache.SetDefault(name, hosts)
	return hosts, nil
}

func (d *k8sDiscovery) Find(svc string) (host string, err error) {
	d.log.Debugf("requested service %s", svc)

	hosts, err := d.getEndpoints(svc)
	if err != nil {
		return "", err
	}

	d.log.Debugf("found %d hosts", len(hosts))

	if len(hosts) == 0 {
		return "", discovery.ErrServiceNotRegistered
	}

	return d.opts.Selection(hosts), nil
}
