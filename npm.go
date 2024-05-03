package main

import (
	"encoding/json"
	"fmt"
	"github.com/jsumners/go-rfc3339"
	"net/http"
	"strings"
)

type NpmClient struct {
	baseUrl string
	http    *http.Client
}

type NpmPackage struct {
	Version string `json:"version"`
}

type NpmDetailedPackage struct {
	// Versions is a map where the key is a version string and the value is
	// the fully rendered `package.json` for that version as it is stored in
	// the registry.
	Versions map[string]any `json:"versions"`

	// Time is a map where the key is a version string and the value is
	// the date and time that version was published to the registry.
	Time map[string]rfc3339.DateTime `json:"time"`
}

type NpmClientOption func(*NpmClient)

func NewNpmClient(options ...NpmClientOption) *NpmClient {
	client := &NpmClient{
		baseUrl: "https://registry.npmjs.com",
		http:    http.DefaultClient,
	}

	for _, opt := range options {
		opt(client)
	}

	return client
}

func WithBaseUrl(url string) NpmClientOption {
	if strings.HasSuffix(url, "/") {
		url = url[0 : len(url)-1]
	}
	return func(client *NpmClient) {
		client.baseUrl = url
	}
}

func WithHttpClient(c *http.Client) NpmClientOption {
	return func(client *NpmClient) {
		client.http = c
	}
}

// GetDetailedInfo gets the full detailed information about a package from the
// NPM registry.
func (nc *NpmClient) GetDetailedInfo(packageName string) (*NpmDetailedPackage, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/%s", nc.baseUrl, packageName),
		nil,
	)
	if err != nil {
		return nil, err
	}

	res, err := nc.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("expected response code 200 but got %d: %s", res.StatusCode, res.Status)
	}

	var body NpmDetailedPackage
	err = json.NewDecoder(res.Body).Decode(&body)
	if err != nil {
		return nil, err
	}

	return &body, nil
}

// GetLatest retrieves the latest version string for the given package.
func (nc *NpmClient) GetLatest(packageName string) (string, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/%s/latest", nc.baseUrl, packageName),
		nil,
	)
	if err != nil {
		return "", err
	}

	res, err := nc.http.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("expected response code 200 but got %d: %s", res.StatusCode, res.Status)
	}

	var body NpmPackage
	err = json.NewDecoder(res.Body).Decode(&body)
	if err != nil {
		return "", err
	}

	return body.Version, nil
}
