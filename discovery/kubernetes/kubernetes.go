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
	"k8s.io/apimachinery/pkg/watch"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type podHosts = map[string]struct{}

type k8sDiscovery struct {
	log       logger.Logger
	opts      *k8sOptions
	podClient corev1.PodInterface

	hostsLock sync.Mutex
	hosts     map[string]podHosts
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

	d.podClient = cl.Pods(d.opts.Namespace)
	resVer, err := d.refreshAll()
	if err != nil {
		return fmt.Errorf("list pods: %w", err)
	}

	d.log.Debugf("resource version is %s", resVer)

	w, err := d.podClient.Watch(context.Background(), v1.ListOptions{ResourceVersion: resVer})
	if err != nil {
		return fmt.Errorf("start watch: %w", err)
	}

	go d.watch(w)
	go d.startRefresh()

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
			} else {
				d.unregister(pod)
			}

		case watch.Deleted:
			if d.unregister(pod) {
				d.log.Debugf("unregistered service '%s' ip %s", svc, pod.Status.PodIP)
			}
		}
	}
}

func (d *k8sDiscovery) startRefresh() {
	t := time.Tick(d.opts.RefreshInterval)

	for range t {
		d.refreshAll()
	}
}

func (d *k8sDiscovery) refreshAll() (resVer string, err error) {
	d.hostsLock.Lock()
	defer d.hostsLock.Unlock()

	d.log.Debugf("refreshing all pods")

	d.hosts = make(map[string]podHosts)

	podlist, err := d.podClient.List(context.Background(), v1.ListOptions{LabelSelector: "mice"})
	if err != nil {
		return "", fmt.Errorf("list pods: %w", err)
	}

	for _, pod := range podlist.Items {
		ok := d.register(&pod)
		d.log.Debugf("added pod %s: %t", pod.Name, ok)
	}

	d.log.Debugf("registered %d pods, resource version is %s", len(d.hosts), podlist.ResourceVersion)

	return podlist.ResourceVersion, nil
}

func (d *k8sDiscovery) register(pod *otherv1.Pod) (added bool) {
	d.hostsLock.Lock()
	defer d.hostsLock.Unlock()

	d.log.Debugf("attempting to register pod %s", pod.Name)

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
	d.hostsLock.Lock()
	defer d.hostsLock.Unlock()

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

	return "", discovery.ErrServiceNotRegistered
}
