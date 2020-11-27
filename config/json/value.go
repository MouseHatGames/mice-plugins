package json

import "time"

type jsonValue struct {
	v interface{}
}

var empty = &jsonValue{}

func (j *jsonValue) Bool(def bool) bool {
	if c, ok := j.v.(bool); ok {
		return c
	}
	return def
}

func (j *jsonValue) Int(def int) int {
	if c, ok := j.v.(int); ok {
		return c
	}
	return def
}

func (j *jsonValue) String(def string) string {
	if c, ok := j.v.(string); ok {
		return c
	}
	return def
}

func (j *jsonValue) Float64(def float64) float64 {
	if c, ok := j.v.(float64); ok {
		return c
	}
	return def
}

func (j *jsonValue) Duration(def time.Duration) time.Duration {
	if c, ok := j.v.(string); ok {
		dur, err := time.ParseDuration(c)
		if err != nil {
			return def
		}
		return dur
	}
	return def
}

func (j *jsonValue) Strings(def []string) []string {
	if c, ok := j.v.([]string); ok {
		return c
	}
	return def
}

func (j *jsonValue) StringMap(def map[string]string) map[string]string {
	if c, ok := j.v.(map[string]string); ok {
		return c
	}
	return def
}

func (j *jsonValue) Scan(val interface{}) error {
	//TODO: Implement
	panic("not implemented")
}
