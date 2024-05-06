package main

import (
	"errors"
	"github.com/jsumners/go-rfc3339"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type RequestErrorTripper struct{}

func (ret *RequestErrorTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("bad request")
}

func Test_WithBaseUrl(t *testing.T) {
	npm := NewNpmClient(WithBaseUrl("http://127.0.0.1/"))
	assert.Equal(t, "http://127.0.0.1", npm.baseUrl)
}

func Test_GetDetailedInfo(t *testing.T) {
	t.Run("returns error for bad request construction", func(t *testing.T) {
		npm := NewNpmClient(WithBaseUrl("http://127.0.0.1"))
		result, err := npm.GetDetailedInfo("foo#%0x24")
		assert.Empty(t, result)
		assert.ErrorContains(t, err, "invalid URL escape")
	})

	t.Run("returns error for server error", func(t *testing.T) {
		client := &http.Client{
			Transport: &RequestErrorTripper{},
		}
		npm := NewNpmClient(WithBaseUrl("http://127.0.0.1"), WithHttpClient(client))

		result, err := npm.GetDetailedInfo("anything")
		assert.Empty(t, result)
		assert.ErrorContains(t, err, "bad request")
	})

	t.Run("returns error for bad payload", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			io.WriteString(res, `{"foo":"bar"`)
		}))
		npm := NewNpmClient(WithBaseUrl(ts.URL))

		result, err := npm.GetDetailedInfo("anything")
		assert.Empty(t, result)
		assert.ErrorContains(t, err, "unexpected EOF")
	})

	t.Run("handles error codes", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(500)
			io.WriteString(res, `failed`)
		}))
		npm := NewNpmClient(WithBaseUrl(ts.URL))

		result, err := npm.GetDetailedInfo("anything")
		assert.Empty(t, result)
		assert.ErrorContains(t, err, "expected response code 200 but got 500: 500")
	})

	t.Run("returns a success response", func(t *testing.T) {
		payload := `{
			"versions": {
				"1.0.0": {}
			},
			"time": {
				"1.0.0": "2024-05-03T13:00:00.000-04:00"
			}
		}`
		ts := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			io.WriteString(res, payload)
		}))
		npm := NewNpmClient(WithBaseUrl(ts.URL))

		dt, _ := rfc3339.NewDateTimeFromString("2024-05-03T13:00:00.000-04:00")
		expected := &NpmDetailedPackage{
			Versions: map[string]any{"1.0.0": make(map[string]any)},
			Time: map[string]rfc3339.DateTime{
				"1.0.0": dt,
			},
		}
		result, err := npm.GetDetailedInfo("anything")
		assert.Nil(t, err)
		assert.Equal(t, expected, result)
	})
}

func Test_GetLatest(t *testing.T) {
	t.Run("returns error for bad request construction", func(t *testing.T) {
		npm := NewNpmClient(WithBaseUrl("http://127.0.0.1"))
		result, err := npm.GetLatest("foo#%0x24")
		assert.Empty(t, result)
		assert.ErrorContains(t, err, "invalid URL escape")
	})

	t.Run("returns error for server error", func(t *testing.T) {
		client := &http.Client{
			Transport: &RequestErrorTripper{},
		}
		npm := NewNpmClient(WithBaseUrl("http://127.0.0.1"), WithHttpClient(client))

		result, err := npm.GetLatest("anything")
		assert.Empty(t, result)
		assert.ErrorContains(t, err, "bad request")
	})

	t.Run("returns error for bad payload", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			io.WriteString(res, `{"foo":"bar"`)
		}))
		npm := NewNpmClient(WithBaseUrl(ts.URL))

		result, err := npm.GetLatest("anything")
		assert.Empty(t, result)
		assert.ErrorContains(t, err, "unexpected EOF")
	})

	t.Run("handles error codes", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(500)
			io.WriteString(res, `failed`)
		}))
		npm := NewNpmClient(WithBaseUrl(ts.URL))

		result, err := npm.GetLatest("anything")
		assert.Empty(t, result)
		assert.ErrorContains(t, err, "expected response code 200 but got 500: 500")
	})

	t.Run("returns a success response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			io.WriteString(res, `{"version":"1.0.0"}`)
		}))
		npm := NewNpmClient(WithBaseUrl(ts.URL))

		result, err := npm.GetLatest("anything")
		assert.Nil(t, err)
		assert.Equal(t, "1.0.0", result)
	})
}
