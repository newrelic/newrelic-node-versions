package main

import (
	"context"
	"errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
)

var nilLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

type logCollector struct {
	logs []string
}

func (lc *logCollector) Write(p []byte) (int, error) {
	lc.logs = append(lc.logs, string(p))
	return len(p), nil
}

// errorReader is an implementation of [io.Reader] that always
// returns an error on read.
type errorReader struct{}

func (er *errorReader) Read([]byte) (int, error) {
	return 0, errors.New("boom")
}

func Test_buildLogger(t *testing.T) {
	t.Run("returns debug level logger", func(t *testing.T) {
		logger := buildLogger(true)
		assert.Equal(t, true, logger.Handler().Enabled(context.TODO(), slog.LevelError))
		assert.Equal(t, true, logger.Handler().Enabled(context.TODO(), slog.LevelDebug))
	})

	t.Run("returns standard logger", func(t *testing.T) {
		logger := buildLogger(false)
		assert.Equal(t, true, logger.Handler().Enabled(context.TODO(), slog.LevelInfo))
		assert.Equal(t, true, logger.Handler().Enabled(context.TODO(), slog.LevelError))
		assert.Equal(t, false, logger.Handler().Enabled(context.TODO(), slog.LevelDebug))
	})
}

func Test_processVersionedTestDirs(t *testing.T) {
	t.Run("parses a versioned test dir", func(t *testing.T) {
		collector := &logCollector{}
		logger := slog.New(slog.NewJSONHandler(collector, &slog.HandlerOptions{Level: slog.LevelError}))
		testDirs := []string{"testdata/versioned"}

		releaseData := processVersionedTestDirs(testDirs, logger)
		assert.Equal(t, 0, len(collector.logs))
		assert.Equal(t, 3, len(releaseData))
	})
}

func Test_readPackageJson(t *testing.T) {
	t.Run("errors for bad file reader", func(t *testing.T) {
		data, err := readPackageJson(&errorReader{})
		assert.Nil(t, data)
		assert.ErrorContains(t, err, "boom")
	})

	t.Run("errors for bad package data", func(t *testing.T) {
		file, err := os.Open(path.Join("testdata", "bad-versioned-package.json"))
		require.Nil(t, err)
		data, err := readPackageJson(file)
		assert.Nil(t, data)
		assert.ErrorContains(t, err, "cannot unmarshal object into")
	})

	t.Run("reads a good file", func(t *testing.T) {
		file, err := os.Open(path.Join("testdata", "latest-version.json"))
		require.Nil(t, err)

		expected := &VersionedTestPackageJson{
			Name: "latest-range",
			Targets: []Target{
				{Name: "foo", MinAgentVersion: "1.0.0"},
			},
			Version: "",
			Private: false,
			Tests: []TestDescription{
				{
					Supported: true,
					Comment:   "",
					Engines:   EnginesBlock{},
					Dependencies: DependenciesBlock{
						"foo": DependencyBlock{
							Versions: "latest",
							Samples:  0,
						},
					},
					Files: nil,
				},
			},
		}
		data, err := readPackageJson(file)
		assert.Nil(t, err)
		assert.Equal(t, expected, data)
	})
}

func Test_cloneRepos(t *testing.T) {
	origFS := appFS
	t.Cleanup(func() {
		appFS = origFS
	})

	t.Run("clones multiple repos", func(t *testing.T) {
		appFS = afero.NewMemMapFs()
		repos := []nrRepo{
			{url: "testdata/bare-repo.git", testPath: "a"},
			{url: "testdata/bare-repo.git", testPath: "b"},
		}
		results := cloneRepos(repos, nilLogger)
		assert.Equal(t, 2, len(results))
		for _, result := range results {
			assert.Nil(t, result.Error)
			assert.Equal(t, true, strings.ContainsAny(result.TestDirectory, "ab"))
			assert.Equal(t, true, strings.Contains(result.Directory, "/newrelic"))
		}
	})
}

func Test_cloneRepo(t *testing.T) {
	origFS := appFS
	t.Cleanup(func() {
		appFS = origFS
	})

	t.Run("returns repo info if local repo dir provided", func(t *testing.T) {
		repo := nrRepo{
			repoDir:  "/foo/bar",
			testPath: "versioned/tests",
		}
		result := cloneRepo(repo, nilLogger)
		assert.Nil(t, result.Error)
		assert.Equal(t, result.Directory, "/foo/bar")
		assert.Equal(t, result.TestDirectory, "versioned/tests")
	})

	t.Run("returns error from creating temp dir", func(t *testing.T) {
		appFS = afero.NewReadOnlyFs(afero.NewMemMapFs())
		repo := nrRepo{
			url:      "https://git.example.com/foo",
			branch:   "main",
			testPath: "test/versioned",
		}
		result := cloneRepo(repo, nilLogger)
		assert.NotNil(t, result.Error)
		assert.ErrorContains(t, result.Error, "failed to create temporary directory")
	})

	t.Run("returns error for bad remote response", func(t *testing.T) {
		appFS = afero.NewMemMapFs()
		ts := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(500)
			res.Write([]byte("bad"))
		}))
		defer ts.Close()

		repo := nrRepo{
			url:      ts.URL,
			branch:   "main",
			testPath: "test/versioned",
		}
		result := cloneRepo(repo, nilLogger)
		assert.NotNil(t, result.Error)
		assert.ErrorContains(t, result.Error, "unexpected client error")
	})

	t.Run("clones repo into temp dir", func(t *testing.T) {
		appFS = afero.NewMemMapFs()
		repo := nrRepo{
			url:      "testdata/bare-repo.git",
			branch:   "main",
			testPath: "test/versioned",
		}
		result := cloneRepo(repo, nilLogger)
		assert.Nil(t, result.Error)
		assert.Equal(t, true, strings.Contains(result.Directory, "/newrelic"))
		assert.Equal(t, "test/versioned", result.TestDirectory)
	})
}

func Test_releaseDataSorter(t *testing.T) {
	a := ReleaseData{Name: "same"}
	b := ReleaseData{Name: "same"}
	assert.Equal(t, 0, releaseDataSorter(a, b))

	a = ReleaseData{Name: "second"}
	b = ReleaseData{Name: "first"}
	assert.Equal(t, 1, releaseDataSorter(a, b))

	a = ReleaseData{Name: "first"}
	b = ReleaseData{Name: "second"}
	assert.Equal(t, -1, releaseDataSorter(a, b))
}

func Test_pruneData(t *testing.T) {
	// Short circuits for a single element.
	input := []ReleaseData{
		{Name: "foo", MinSupportedVersion: "1.0.0"},
	}
	expected := []ReleaseData{
		{Name: "foo", MinSupportedVersion: "1.0.0"},
	}
	assert.Equal(t, expected, pruneData(input))

	// Drops a literal duplicate.
	input = []ReleaseData{
		{Name: "foo", MinSupportedVersion: "1.0.0"},
		{Name: "foo", MinSupportedVersion: "1.0.0"},
	}
	expected = []ReleaseData{
		{Name: "foo", MinSupportedVersion: "1.0.0"},
	}
	assert.Equal(t, expected, pruneData(input))

	// Picks first one.
	input = []ReleaseData{
		{Name: "foo", MinSupportedVersion: "1.0.0"},
		{Name: "foo", MinSupportedVersion: "2.0.0"},
	}
	expected = []ReleaseData{
		{Name: "foo", MinSupportedVersion: "1.0.0"},
	}
	assert.Equal(t, expected, pruneData(input))

	// Picks second one.
	input = []ReleaseData{
		{Name: "foo", MinSupportedVersion: "2.0.0"},
		{Name: "foo", MinSupportedVersion: "1.0.0"},
	}
	expected = []ReleaseData{
		{Name: "foo", MinSupportedVersion: "1.0.0"},
	}
	assert.Equal(t, expected, pruneData(input))

	// All-in-one.
	input = []ReleaseData{
		{Name: "foo", MinSupportedVersion: "2.0.0"},
		{Name: "foo", MinSupportedVersion: "1.0.0"},
		{Name: "bar", MinSupportedVersion: "1.0.0"},
		{Name: "baz", MinSupportedVersion: "3.0.0"},
	}
	expected = []ReleaseData{
		{Name: "foo", MinSupportedVersion: "1.0.0"},
		{Name: "bar", MinSupportedVersion: "1.0.0"},
		{Name: "baz", MinSupportedVersion: "3.0.0"},
	}
	assert.Equal(t, expected, pruneData(input))
}
