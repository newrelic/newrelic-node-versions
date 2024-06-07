package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_createAndParseFlags(t *testing.T) {
	t.Run("defaults are as expected", func(t *testing.T) {
		err := createAndParseFlags([]string{"ignored"})
		expected := appFlags{
			aiCompatJsonFile: "",
			noExternals:      false,
			startMarker:      "{/* begin: compat-table */}",
			endMarker:        "{/* end: compat-table */}",
		}
		assert.Nil(t, err)
		assert.Equal(t, expected, flags)
	})
}
