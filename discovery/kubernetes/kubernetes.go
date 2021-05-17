package kubernetes

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/MouseHatGames/mice/discovery"
	"github.com/MouseHatGames/mice/logger"
	"github.com/MouseHatGames/mice/options"
	otherv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type endpoints = []string

type k8sDiscovery struct {
	log            logger.Logger
	opts           *k8sOptions
	endpointClient corev1.EndpointsInterface

	lastRefresh time.Time
	hostsLock   sync.Mutex
	hosts       map[string]endpoints
}

func Discovery(opts ...K8sOption) options.Option {
	return func(o *options.Options) {
		k8opt := &k8sOptions{
			Namespace:       "default",
			RefreshInterval: 30 * time.Second,
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
			hosts: make(map[string]endpoints),
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

func (d *k8sDiscovery) refreshAll() error {
	d.hostsLock.Lock()
	defer d.hostsLock.Unlock()

	d.log.Debugf("refreshing all endpoints")

	d.hosts = make(map[string]endpoints)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	endpoints, err := d.endpointClient.List(ctx, v1.ListOptions{})
	if err != nil {
		return fmt.Errorf("list pods: %w", err)
	}

	for _, ep := range endpoints.Items {
		name, ok := d.getEndpointHosts(&ep)
		if !ok {
			continue
		}

		for _, subset := range ep.Subsets {
			for _, addr := range subset.Addresses {
				d.hosts[name] = append(d.hosts[name], addr.IP)
			}
		}

		d.log.Debugf("added endpoint %s: %t", ep.Name, ok)
	}

	d.log.Debugf("registered %d endpoints", len(d.hosts))
	d.lastRefresh = time.Now()

	return nil
}

func (d *k8sDiscovery) refreshIfStale() {
	if time.Now().Sub(d.lastRefresh) <= d.opts.RefreshInterval {
		return
	}

	d.log.Debugf("stale, refreshing")

	go func() {
		err := d.refreshAll()
		if err != nil {
			d.log.Errorf("failed to refresh stale: %s", err)
		}
	}()
}

func (d *k8sDiscovery) getEndpointHosts(eps *otherv1.Endpoints) (string, bool) {
	svc, ok := eps.Labels["mice"]

	if !ok {
		return "", false
	}

	_, ok = d.hosts[svc]
	if !ok {
		d.hosts[svc] = make(endpoints, 0, 3)
	}

	return svc, true
}

func (d *k8sDiscovery) Find(svc string) (host string, err error) {
	d.log.Debugf("requested service %s", svc)

	d.refreshIfStale()

	hosts := d.hosts[svc]

	d.log.Debugf("found %d hosts", len(hosts))

	if len(hosts) == 0 {
		return "", discovery.ErrServiceNotRegistered
	}

	return d.opts.Selection(hosts), nil
}
