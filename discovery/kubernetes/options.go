package kubernetes

type k8sOptions struct {
	Namespace string
}

type k8sOption func(*k8sOptions)

func Namespace(ns string) k8sOption {
	return func(o *k8sOptions) {
		o.Namespace = ns
	}
}
