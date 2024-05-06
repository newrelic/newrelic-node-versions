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

const onlySamplesInBlock = `{
	"foo": {
		"samples": "5"
	}
}`

func Test_TestDescription(t *testing.T) {
	t.Run("unmarshals null value", func(t *testing.T) {
		var td TestDescription
		err := json.Unmarshal([]byte("null"), &td)
		assert.Nil(t, err)
		assert.Empty(t, td)
	})

	t.Run("supported defaults true", func(t *testing.T) {
		var td TestDescription
		err := json.Unmarshal([]byte(`{}`), &td)
		assert.Nil(t, err)
		assert.Equal(t, true, td.Supported)
	})

	t.Run("unmarshals supported true", func(t *testing.T) {
		var td TestDescription
		err := json.Unmarshal([]byte(`{"supported":true}`), &td)
		assert.Nil(t, err)
		assert.Equal(t, true, td.Supported)
	})

	t.Run("unmarshals supported false", func(t *testing.T) {
		var td TestDescription
		err := json.Unmarshal([]byte(`{"supported":false}`), &td)
		assert.Nil(t, err)
		assert.Equal(t, false, td.Supported)
	})

	t.Run("unmarshals comment", func(t *testing.T) {
		var td TestDescription
		err := json.Unmarshal([]byte(`{"comment":"foo"}`), &td)
		assert.Nil(t, err)
		assert.Equal(t, "foo", td.Comment)
	})

	t.Run("returns error for bad comment", func(t *testing.T) {
		var td TestDescription
		err := json.Unmarshal([]byte(`{"comment":42}`), &td)
		assert.ErrorContains(t, err, "cannot unmarshal number into")
	})

	t.Run("unmarshals engines", func(t *testing.T) {
		var td TestDescription
		err := json.Unmarshal([]byte(`{"engines":{"node":"20"}}`), &td)
		assert.Nil(t, err)
		assert.Equal(t, "20", td.Engines.Node)
	})

	t.Run("returns error for bad engines", func(t *testing.T) {
		var td TestDescription
		err := json.Unmarshal([]byte(`{"engines":42}`), &td)
		assert.ErrorContains(t, err, "cannot unmarshal number into")
	})

	t.Run("unmarsals dependencies", func(t *testing.T) {
		var td TestDescription
		err := json.Unmarshal([]byte(`{"dependencies":{"foo":{"versions":"1.0.0"}}}`), &td)
		assert.Nil(t, err)
		assert.Equal(t, "1.0.0", td.Dependencies["foo"].Versions)
	})

	t.Run("returns error for bad dependencies", func(t *testing.T) {
		var td TestDescription
		err := json.Unmarshal([]byte(`{"dependencies":42}`), &td)
		assert.ErrorContains(t, err, "cannot unmarshal number")
	})

	t.Run("unmarshals files", func(t *testing.T) {
		var td TestDescription
		err := json.Unmarshal([]byte(`{"files":["foo"]}`), &td)
		assert.Nil(t, err)
		assert.Equal(t, "foo", td.Files[0])
	})

	t.Run("returns erro for bad files", func(t *testing.T) {
		var td TestDescription
		err := json.Unmarshal([]byte(`{"files":{"foo":"bar"}}`), &td)
		assert.ErrorContains(t, err, "cannot unmarshal object into")
	})
}

func Test_DependenciesBlock(t *testing.T) {
	t.Run("unmarshals null value", func(t *testing.T) {
		var block DependenciesBlock
		err := json.Unmarshal([]byte("null"), &block)
		assert.Nil(t, err)
		assert.Empty(t, block)
	})

	t.Run("returns error for bad block", func(t *testing.T) {
		var block DependenciesBlock
		err := json.Unmarshal([]byte(`[42]`), &block)
		assert.ErrorContains(t, err, "cannot unmarshal array")
	})

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

	t.Run("unmarshals block with only samples", func(t *testing.T) {
		var block DependenciesBlock
		err := json.Unmarshal([]byte(onlySamplesInBlock), &block)
		assert.Empty(t, block)
		assert.ErrorContains(t, err, "failed to parse dependency block: missing versions property")
	})
}

func Test_DependencyBlock(t *testing.T) {
	t.Run("unmarshals null value", func(t *testing.T) {
		var block DependencyBlock
		err := json.Unmarshal([]byte("null"), &block)
		assert.Nil(t, err)
		assert.Empty(t, block)
	})

	t.Run("unmarshals string samples", func(t *testing.T) {
		var block DependencyBlock
		err := json.Unmarshal([]byte(`{"versions":"1.0.0","samples":"5"}`), &block)
		assert.Nil(t, err)
		assert.Equal(t, 5, block.Samples)
	})

	t.Run("unmarshals int samples", func(t *testing.T) {
		var block DependencyBlock
		err := json.Unmarshal([]byte(`{"versions":"1.0.0","samples":5}`), &block)
		assert.Nil(t, err)
		assert.Equal(t, 5, block.Samples)
	})

	t.Run("returns error if missing versions", func(t *testing.T) {
		var block DependencyBlock
		err := json.Unmarshal([]byte(onlySamplesInBlock), &block)
		assert.ErrorContains(t, err, "missing versions property")
	})
}
