package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func Test_ParsePackage_Elastic(t *testing.T) {
	var pkgJson VersionedTestPackageJson
	file, err := os.ReadFile("testdata/versioned/elastic/package.json")
	require.Nil(t, err)
	err = json.Unmarshal(file, &pkgJson)
	require.Nil(t, err)

	t.Run("parses to correct representation", func(t *testing.T) {
		expected := &PkgInfo{
			Name:       "@elastic/elasticsearch",
			MinVersion: "7.16.0",
		}
		found, err := parsePackage(pkgJson)
		assert.Nil(t, err)
		assert.Equal(t, expected, found)
	})
}

func Test_ParsePackage_Langchain(t *testing.T) {
	var pkgJson VersionedTestPackageJson
	file, err := os.ReadFile("testdata/versioned/langchain/package.json")
	require.Nil(t, err)
	err = json.Unmarshal(file, &pkgJson)
	require.Nil(t, err)

	t.Run("parses to correct representation", func(t *testing.T) {
		expected := &PkgInfo{
			Name:       "@langchain/core",
			MinVersion: "0.1.17",
		}
		found, err := parsePackage(pkgJson)
		assert.Nil(t, err)
		assert.Equal(t, expected, found)
	})
}
