package kubernetes

import (
	"context"
	"fmt"

	"github.com/MouseHatGames/mice/discovery"
	"github.com/MouseHatGames/mice/logger"
	"github.com/MouseHatGames/mice/options"
	otherv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type podHosts = map[string]struct{}

type k8sDiscovery struct {
	log  logger.Logger
	opts *k8sOptions

	hosts map[string]podHosts
}

func Discovery(opts ...K8sOption) options.Option {
	return func(o *options.Options) {
		k8opt := &k8sOptions{
			Namespace: "default",
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
			hosts: make(map[string]podHosts),
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

	pods := cl.Pods(d.opts.Namespace)

	podlist, err := pods.List(context.Background(), v1.ListOptions{LabelSelector: "mice"})
	if err != nil {
		return fmt.Errorf("list pods: %w", err)
	}

	for _, pod := range podlist.Items {
		d.register(&pod)
	}

	w, err := pods.Watch(context.Background(), v1.ListOptions{})
	if err != nil {
		return fmt.Errorf("start watch: %w", err)
	}

	go d.watch(w)

	return nil
}

func (d *k8sDiscovery) watch(w watch.Interface) {
	d.log.Debugf("starting watcher")

	for ev := range w.ResultChan() {
		d.log.Debugf("received event: %s", ev.Type)

		pod := ev.Object.(*otherv1.Pod)
		svc, ok := pod.Labels["mice"]

		if !ok {
			continue
		}

		switch ev.Type {
		case watch.Modified, watch.Added:
			if pod.Status.Phase == otherv1.PodRunning {
				d.register(pod)
			}

		case watch.Deleted:
			if d.unregister(pod) {
				d.log.Debugf("unregistered service '%s' ip %s", svc, pod.Status.PodIP)
			}
		}
	}
}

func (d *k8sDiscovery) register(pod *otherv1.Pod) (added bool) {
	ip := pod.Status.PodIP
	if ip == "" {
		return false
	}

	hosts := d.getPodHosts(pod)
	if hosts == nil {
		return false
	}

	_, added = hosts[ip]
	hosts[ip] = struct{}{}

	if added {
		d.log.Debugf("registered new service '%s' pod: %s ip: %s", pod.Labels["mice"], pod.Name, pod.Status.PodIP)
	}

	return added
}

func (d *k8sDiscovery) unregister(pod *otherv1.Pod) (removed bool) {
	ip := pod.Status.PodIP
	if ip == "" {
		return false
	}

	hosts := d.getPodHosts(pod)
	if hosts == nil {
		return false
	}

	_, removed = hosts[ip]
	if removed {
		delete(hosts, ip)
	}

	return removed
}

func (d *k8sDiscovery) getPodHosts(pod *otherv1.Pod) podHosts {
	svc, ok := pod.Labels["mice"]

	if !ok {
		return nil
	}

	m, ok := d.hosts[svc]
	if !ok {
		m = make(podHosts)
		d.hosts[svc] = m
	}

	return m
}

func (d *k8sDiscovery) Find(svc string) (host string, err error) {
	d.log.Debugf("requested service %s", svc)

	hosts := d.hosts[svc]

	d.log.Debugf("found %d hosts", len(hosts))

	// Return first host if the map is not empty
	for k := range hosts {
		d.log.Debugf("found host %s for %s", k, svc)
		return k, nil
	}

	d.log.Errorf("no host found for %s", svc)
	return "", discovery.ErrServiceNotRegistered
}
