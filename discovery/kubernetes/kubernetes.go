package kubernetes

import (
	"context"
	"fmt"
	"time"

	"github.com/MouseHatGames/mice/discovery"
	"github.com/MouseHatGames/mice/options"
	"github.com/op/go-logging"
	otherv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type k8sDiscovery struct {
	log  *logging.Logger
	opts *k8sOptions

	hosts map[string]map[string]time.Time
}

func Discovery(opts ...k8sOption) options.Option {
	return func(o *options.Options) {
		k8opt := &k8sOptions{
			Namespace: "default",
		}

		for _, opt := range opts {
			opt(k8opt)
		}

		o.Discovery = &k8sDiscovery{
			opts: k8opt,
			log:  logging.MustGetLogger("k8s"),
		}
	}
}

func (d *k8sDiscovery) Start() error {
	d.log.Debug("getting in-cluster config")

	config, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("get in-cluster config: %w", err)
	}

	d.log.Debug("creating client")

	cl, err := corev1.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("get clientset: %w", err)
	}

	pods := cl.Pods(d.opts.Namespace)
	w, err := pods.Watch(context.Background(), v1.ListOptions{})
	if err != nil {
		return fmt.Errorf("start watch: %w", err)
	}

	go d.watch(w)

	return nil
}

func (d *k8sDiscovery) watch(w watch.Interface) {
	d.log.Debug("starting watcher")

	for ev := range w.ResultChan() {
		d.log.Debugf("received event: %s", ev.Type)

		pod := ev.Object.(*otherv1.Pod)
		svc, ok := pod.Labels["mice"]

		if !ok {
			continue
		}

		ip := pod.Status.PodIP
		hosts := d.getHosts(svc)

		switch ev.Type {
		case watch.Modified:
			if ip != "" {
				d.log.Debugf("registered new service '%s' ip: %", svc, ip)
				hosts[ip] = time.Now()
			}

		case watch.Deleted:
			d.log.Debugf("deleted service '%s' ip %s", svc, ip)
			delete(hosts, ip)
		}
	}
}

func (d *k8sDiscovery) getHosts(svc string) map[string]time.Time {
	m, ok := d.hosts[svc]
	if !ok {
		m = make(map[string]time.Time)
		d.hosts[svc] = m
	}

	return m
}

func (d *k8sDiscovery) Find(svc string) (host string, err error) {
	hosts := d.hosts[svc]

	// Return first host if the map is not empty
	for k := range hosts {
		return k, nil
	}

	return "", discovery.ErrServiceNotRegistered
}
