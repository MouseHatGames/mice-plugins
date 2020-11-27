package json

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/MouseHatGames/mice/config"
	"github.com/MouseHatGames/mice/options"
)

var ErrNotMap = errors.New("tried to index a non-map value")
var ErrKeyNotFound = errors.New("key not found on map")

func Config(file string) options.Option {
	return func(o *options.Options) {
		c, err := newConfig(file)
		if err != nil {
			panic(fmt.Sprintf("failed to read json config file: %s", err))
		}

		o.Config = c
	}
}

type jsonConfig struct {
	data map[string]interface{}
}

func newConfig(file string) (*jsonConfig, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(f).Decode(&data); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}

	return &jsonConfig{
		data: data,
	}, nil
}

func (j *jsonConfig) Get(path ...string) config.Value {
	var d interface{} = j.data
	for _, part := range path {
		m, ok := d.(map[string]interface{})
		if !ok {
			return empty
		}

		d, ok = m[part]
		if !ok {
			return empty
		}
	}

	return &jsonValue{d}
}

func (j *jsonConfig) Del(path ...string) error {
	var d interface{} = j.data
	for i, part := range path {
		m, ok := d.(map[string]interface{})
		if !ok {
			return ErrNotMap
		}

		// If we are on the last part of the path, delete the key
		if i == len(path)-1 {
			delete(m, part)
		} else {
			d = m[part]
		}
	}

	return nil
}

func (j *jsonConfig) Set(val interface{}, path ...string) error {
	var d interface{} = j.data
	for i, part := range path {
		m, ok := d.(map[string]interface{})
		if !ok {
			return ErrNotMap
		}

		// If we are on the last part of the path, set the key
		if i == len(path)-1 {
			m[part] = val
		} else {
			d = m[part]
		}
	}

	return nil
}
