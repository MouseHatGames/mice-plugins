package dns

import (
	"fmt"
	"net"

	"github.com/MouseHatGames/mice/options"
)

type dnsDiscovery struct {
	prefix, suffix string
}

type Option func(*dnsDiscovery)

func Prefix(p string) Option {
	return func(dd *dnsDiscovery) {
		dd.prefix = p
	}
}

func Suffix(s string) Option {
	return func(dd *dnsDiscovery) {
		dd.suffix = s
	}
}

func Discovery(opts ...Option) options.Option {
	return func(o *options.Options) {
		dd := &dnsDiscovery{}

		for _, opt := range opts {
			opt(dd)
		}

		o.Discovery = dd
	}
}

func (d *dnsDiscovery) Find(svc string) (host string, err error) {
	svc = fmt.Sprintf("%s%s%s", d.prefix, svc, d.suffix)

	ips, err := net.LookupIP(svc)
	if err != nil {
		return "", fmt.Errorf("lookup host: %w", err)
	}

	if len(ips) == 1 {
		return ips[0].String(), nil
	}

	//TODO: Use a selection algorithm
	return "", fmt.Errorf("more than one ip found")
}
