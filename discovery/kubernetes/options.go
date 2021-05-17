package kubernetes

import (
	"math/rand"
	"time"
)

type selection func(hosts []string) string

type k8sOptions struct {
	Namespace string
	Selection selection
	CacheTime time.Duration
}

type K8sOption func(*k8sOptions)

// Namespace sets the Kubernetes namespace that will be watched for services
func Namespace(ns string) K8sOption {
	return func(o *k8sOptions) {
		o.Namespace = ns
	}
}

// CacheTime sets the time that addresses will be cached before being fetched again
func CacheTime(t time.Duration) K8sOption {
	return func(o *k8sOptions) {
		o.CacheTime = t
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
