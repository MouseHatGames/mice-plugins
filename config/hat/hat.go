package hat

import (
	mhat "github.com/MouseHatGames/hat/pkg/client"
	"github.com/MouseHatGames/mice/config"
	"github.com/MouseHatGames/mice/options"
)

func Config(endpoint string) options.Option {
	return func(o *options.Options) {
		o.Config = &hatConfig{
			endpoint: endpoint,
		}
	}
}

type hatConfig struct {
	endpoint string
	hat      mhat.Client
}

func (c *hatConfig) Start() error {
	cl, err := mhat.Dial(c.endpoint)
	if err != nil {
		return err
	}

	c.hat = cl
	return nil
}

func (c *hatConfig) Get(path ...string) config.Value {
	return c.hat.Get(path...)
}

func (c *hatConfig) Del(path ...string) error {
	return c.hat.Del(path...)
}

func (c *hatConfig) Set(val interface{}, path ...string) error {
	return c.hat.Set(val, path...)
}
