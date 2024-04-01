package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func testPkg(t *testing.T, jsonFile string, expected *PkgInfo) {
	var pkgJson VersionedTestPackageJson
	file, err := os.ReadFile(jsonFile)
	require.Nil(t, err)
	err = json.Unmarshal(file, &pkgJson)
	require.Nil(t, err)

	found, err := parsePackage(pkgJson)
	assert.Nil(t, err)
	assert.Equal(t, expected, found)
}

func Test_ParsePackage(t *testing.T) {
	t.Run("handles out of order ranges", func(t *testing.T) {
		testPkg(t, "testdata/out-of-order-ranges.json", &PkgInfo{
			Name:       "foo",
			MinVersion: "1.5.0",
		})
	})

	t.Run("handles @elastic/elasticsearch", func(t *testing.T) {
		testPkg(t, "testdata/versioned/elastic/package.json", &PkgInfo{
			Name:       "@elastic/elasticsearch",
			MinVersion: "7.16.0",
		})
	})

	t.Run("handles @langchain/core", func(t *testing.T) {
		testPkg(t, "testdata/versioned/langchain/package.json", &PkgInfo{
			Name:       "@langchain/core",
			MinVersion: "0.1.17",
		})
	})

	t.Run("handles mongodb", func(t *testing.T) {
		testPkg(t, "testdata/versioned/mongodb/package.json", &PkgInfo{
			Name:       "mongodb",
			MinVersion: "2.1.0",
		})
	})
}
