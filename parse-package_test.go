package main

import (
	"blitznote.com/src/semver/v3"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func testPkg(t *testing.T, jsonFile string, expected []PkgInfo) {
	var pkgJson VersionedTestPackageJson
	file, err := os.ReadFile(jsonFile)
	require.Nil(t, err)
	err = json.Unmarshal(file, &pkgJson)
	require.Nil(t, err)

	found, err := parsePackage(&pkgJson)
	assert.Nil(t, err)
	assert.Equal(t, expected, found)
}

func Test_ParsePackage(t *testing.T) {
	t.Run("handles out of order ranges", func(t *testing.T) {
		testPkg(t, "testdata/out-of-order-ranges.json", []PkgInfo{{
			Name:            "foo",
			MinVersion:      "1.5.0",
			MinAgentVersion: "0.0.0",
		}})
	})

	t.Run("handles @elastic/elasticsearch", func(t *testing.T) {
		testPkg(t, "testdata/versioned/elastic/package.json", []PkgInfo{{
			Name:            "@elastic/elasticsearch",
			MinVersion:      "7.16.0",
			MinAgentVersion: "1.2.3",
		}})
	})

	t.Run("handles @langchain/core", func(t *testing.T) {
		testPkg(t, "testdata/versioned/langchain/package.json", []PkgInfo{{
			Name:            "@langchain/core",
			MinVersion:      "0.1.17",
			MinAgentVersion: "2.1.3",
		}})
	})

	t.Run("handles mongodb", func(t *testing.T) {
		testPkg(t, "testdata/versioned/mongodb/package.json", []PkgInfo{{
			Name:            "mongodb",
			MinVersion:      "2.1.0",
			MinAgentVersion: "1.0.0",
		}})
	})

	t.Run("gets minimum range from ordered ORed ranges", func(t *testing.T) {
		testPkg(t, "testdata/ordered-or-range.json", []PkgInfo{{
			Name:            "foo",
			MinVersion:      "1.0.0",
			MinAgentVersion: "1.0.0",
		}})
	})

	t.Run("gets minimum range from unordered ORed ranges", func(t *testing.T) {
		testPkg(t, "testdata/unordered-or-range.json", []PkgInfo{{
			Name:            "foo",
			MinVersion:      "1.0.0",
			MinAgentVersion: "1.0.0",
		}})
	})

	t.Run("handles `latest` version string", func(t *testing.T) {
		testPkg(t, "testdata/latest-version.json", []PkgInfo{{
			Name:            "foo",
			MinVersion:      "0.0.0",
			MinAgentVersion: "1.0.0",
		}})
	})
}

func Test_normalizeRangeString(t *testing.T) {
	input := ">=2.1 < 4.0.0 "
	expected := ">=2.1 <4.0.0"
	assert.Equal(t, expected, normalizeRangeString(input))

	input = ">= 4.1.4 < 5"
	expected = ">=4.1.4 <5"
	assert.Equal(t, expected, normalizeRangeString(input))
}

func Test_isVersionLess(t *testing.T) {
	range1, _ := semver.NewRange([]byte("*"))
	range2, _ := semver.NewRange([]byte(max_range))
	assert.Equal(t, true, isRangeLower(range1, range2))

	range1, _ = semver.NewRange([]byte("*"))
	range2, _ = semver.NewRange([]byte("<1.0.0"))
	assert.Equal(t, true, isRangeLower(range1, range2))

	range1, _ = semver.NewRange([]byte("1.0.0"))
	range2, _ = semver.NewRange([]byte("<1.0.0"))
	assert.Equal(t, false, isRangeLower(range1, range2))

	range1, _ = semver.NewRange([]byte(">=0.1.0 <1.0.0"))
	range2, _ = semver.NewRange([]byte(max_range))
	assert.Equal(t, false, isRangeLower(range1, range2))
	range1, _ = semver.NewRange([]byte(max_range))
	range2, _ = semver.NewRange([]byte(">=0.1.0 <1.0.0"))
	assert.Equal(t, false, isRangeLower(range1, range2))

	range1, _ = semver.NewRange([]byte(">=0.1.0 <1.0.0"))
	range2, _ = semver.NewRange([]byte("=<1.0.0"))
	assert.Equal(t, false, isRangeLower(range1, range2))

	range1, _ = semver.NewRange([]byte(">=0.1.0 <1.0.0"))
	range2, _ = semver.NewRange([]byte(">2.0.0 <3.0.0"))
	assert.Equal(t, true, isRangeLower(range1, range2))
}
