package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	data := map[string]interface{}{
		"one": map[string]interface{}{
			"two": 2,
		},
	}

	c := &jsonConfig{data: data}

	t.Run("not found", func(t *testing.T) {
		ret := c.Get("foo", "bar")

		assert.Equal(t, empty, ret)
	})
	t.Run("normal", func(t *testing.T) {
		ret := c.Get("one", "two")

		val := ret.(*jsonValue)

		assert.Equal(t, 2, val.v)
	})
}

func TestDel(t *testing.T) {
	t.Run("not a map", func(t *testing.T) {
		data := map[string]interface{}{
			"one": 123,
		}
		c := &jsonConfig{data: data}

		err := c.Del("one", "two")

		assert.Equal(t, ErrNotMap, err)
	})
	t.Run("normal", func(t *testing.T) {
		data := map[string]interface{}{
			"one": map[string]interface{}{
				"two": 2,
			},
		}
		c := &jsonConfig{data: data}

		err := c.Del("one", "two")

		assert.Nil(t, err)
		assert.Len(t, data["one"], 0)
	})
}

func TestSet(t *testing.T) {
	t.Run("not a map", func(t *testing.T) {
		data := map[string]interface{}{
			"one": 123,
		}
		c := &jsonConfig{data: data}

		err := c.Set(42, "one", "two")

		assert.Equal(t, ErrNotMap, err)
	})
	t.Run("normal", func(t *testing.T) {
		inner := map[string]interface{}{
			"two": 2,
		}
		data := map[string]interface{}{
			"one": inner,
		}
		c := &jsonConfig{data: data}

		err := c.Set(42, "one", "two")

		assert.Nil(t, err)
		assert.Equal(t, 42, inner["two"])
	})
}
