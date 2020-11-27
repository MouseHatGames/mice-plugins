package json

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBool(t *testing.T) {
	const ex = true
	v := &jsonValue{ex}

	assert.Equal(t, ex, v.Bool(false))
}

func TestInt(t *testing.T) {
	const ex = 123
	v := &jsonValue{ex}

	assert.Equal(t, ex, v.Int(0))
}

func TestString(t *testing.T) {
	const ex = "nice"
	v := &jsonValue{ex}

	assert.Equal(t, ex, v.String(""))
}

func TestFloat64(t *testing.T) {
	const ex = 1.23
	v := &jsonValue{ex}

	assert.Equal(t, ex, v.Float64(0))
}

func TestDuration(t *testing.T) {
	var ex = 2 * time.Second
	v := &jsonValue{ex.String()}

	assert.Equal(t, ex, v.Duration(0))
}

func TestStrings(t *testing.T) {
	var ex = []string{"one", "two"}
	v := &jsonValue{ex}

	assert.Equal(t, ex, v.Strings(nil))
}

func TestStringMap(t *testing.T) {
	var ex = map[string]string{"one": "1", "two": "2"}
	v := &jsonValue{ex}

	assert.Equal(t, ex, v.StringMap(nil))
}
