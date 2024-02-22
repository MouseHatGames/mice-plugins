package dns

import (
	"fmt"
	"net"

	"github.com/MouseHatGames/mice/options"
)

type dnsDiscovery struct {
}

func Discovery() options.Option {
	return func(o *options.Options) {
		o.Discovery = &dnsDiscovery{}
	}
}

func (d *dnsDiscovery) Find(svc string) (host string, err error) {
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
