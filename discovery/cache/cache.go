package cache

import (
	"time"

	"github.com/MouseHatGames/mice/discovery"
	"github.com/patrickmn/go-cache"
)

type cacheDiscovery struct {
	inner discovery.Discovery
	cache *cache.Cache

	maxAge time.Duration
}

type Option func(*cacheDiscovery)

func MaxAge(d time.Duration) Option {
	return func(cd *cacheDiscovery) {
		cd.maxAge = d
	}
}

func Discovery(inner discovery.Discovery, opts ...Option) discovery.Discovery {
	cd := &cacheDiscovery{
		inner:  inner,
		maxAge: 5 * time.Minute,
	}

	for _, opt := range opts {
		opt(cd)
	}

	cd.cache = cache.New(cd.maxAge, 5*time.Minute)
	return cd
}

func (c *cacheDiscovery) Find(svc string) (host string, err error) {
	if v, ok := c.cache.Get(svc); ok {
		return v.(string), nil
	}

	host, err = c.inner.Find(svc)
	if err != nil {
		return "", err
	}

	c.cache.SetDefault(svc, host)
	return
}
