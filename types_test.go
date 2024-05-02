package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

const basicDependencyBlock = `{
	"foo": "^1.0.0",
	"bar": ">1.1.0 <=1.19.0"
}`

const complicatedDependencyBlock = `{
	"foo": {
		"versions": "^2.0.0",
		"samples": "2"
	},
	"bar": "2.0.0"
}`

const onlyVersionsInBlock = `{
	"foo": {
		"versions": "1.2.3"
	}
}`

func Test_DependenciesBlock(t *testing.T) {
	t.Run("unmarshals basic block", func(t *testing.T) {
		var block DependenciesBlock
		err := json.Unmarshal([]byte(basicDependencyBlock), &block)
		assert.Nil(t, err)
		expected := DependenciesBlock{
			"foo": DependencyBlock{Versions: "^1.0.0"},
			"bar": DependencyBlock{Versions: ">1.1.0 <=1.19.0"},
		}
		assert.Equal(t, expected, block)
	})

	t.Run("unmarshals complicated block", func(t *testing.T) {
		var block DependenciesBlock
		err := json.Unmarshal([]byte(complicatedDependencyBlock), &block)
		assert.Nil(t, err)
		expected := DependenciesBlock{
			"foo": DependencyBlock{Versions: "^2.0.0", Samples: 2},
			"bar": DependencyBlock{Versions: "2.0.0"},
		}
		assert.Equal(t, expected, block)
	})

	t.Run("unmarshals block with only versions", func(t *testing.T) {
		var block DependenciesBlock
		err := json.Unmarshal([]byte(onlyVersionsInBlock), &block)
		assert.Nil(t, err)
		expected := DependenciesBlock{
			"foo": DependencyBlock{Versions: "1.2.3"},
		}
		assert.Equal(t, expected, block)
	})
}
