package kubernetes

import (
	"math/rand"
	"time"
)

type selection func(hosts []string) string

type k8sOptions struct {
	Namespace       string
	Selection       selection
	RefreshInterval time.Duration
}

type K8sOption func(*k8sOptions)

// Namespace sets the Kubernetes namespace that will be watched for services
func Namespace(ns string) K8sOption {
	return func(o *k8sOptions) {
		o.Namespace = ns
	}
}

// RefreshInterval sets the interval at which pods will be refreshed
func RefreshInterval(t time.Duration) K8sOption {
	return func(o *k8sOptions) {
		o.RefreshInterval = t
	}
}

// RoundRobin indicates that services will be selected one after the other
func RoundRobin() K8sOption {
	return func(o *k8sOptions) {
		var counter uint64

		o.Selection = func(hosts []string) string {
			n := counter % uint64(len(hosts))
			counter++

			return hosts[n]
		}
	}
}

// Random indicates that services will be selected randomly
func Random() K8sOption {
	return func(o *k8sOptions) {
		o.Selection = func(hosts []string) string {
			n := rand.Intn(len(hosts))

			return hosts[n]
		}
	}
}
