package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_StringEnumValue(t *testing.T) {
	t.Run("only allows specified values", func(t *testing.T) {
		sev := NewStringEnumValue([]string{"foo", "bar"}, "foo")
		expected := &StringEnumValue{
			allowed: []string{"foo", "bar"},
			value:   "foo",
		}
		assert.Equal(t, expected, sev)

		err := sev.Set("baz")
		assert.ErrorContains(t, err, "baz is not an allowed value")

		err = sev.Set("bar")
		assert.Nil(t, err)
		expected.value = "bar"
		assert.Equal(t, expected, sev)
	})

	t.Run("supports interface methods", func(t *testing.T) {
		sev := NewStringEnumValue([]string{"foo", "bar"}, "foo")
		assert.Equal(t, "string", sev.Type())
		assert.Equal(t, "foo", sev.String())

		err := sev.Set("bar")
		assert.Nil(t, err)
		assert.Equal(t, "bar", sev.String())
	})
}
