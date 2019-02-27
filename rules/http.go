package rules

import (
	"fmt"
	"io"
	"net"
	"net/http"
	nu "net/url"
	"strings"
	"time"
)

var (
	createClient = getNetClient
)

type httpClient interface {
	SetTimeout(time.Duration)
	Get(string) (*http.Response, error)
	Post(string, string, io.Reader) (*http.Response, error)
}

type wrappedHTTPClient struct {
	*http.Client
}

func (c *wrappedHTTPClient) SetTimeout(dur time.Duration) {
	c.Client.Timeout = dur
}

func (c *wrappedHTTPClient) Get(url string) (*http.Response, error) {
	return c.Client.Get(url)
}

func (c *wrappedHTTPClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	return c.Client.Post(url, contentType, body)
}

// This is copied originally from http.Transport:
var transport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}).DialContext,
	MaxIdleConns:          200, // Original value of 100 bumped to 200.
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

func getNetClient(duration time.Duration) httpClient {
	return &wrappedHTTPClient{
		Client: &http.Client{
			Transport: transport,
			Timeout:   duration,
		},
	}
}

func isValidURL(url string) bool {
	if len(url) == 0 {
		return false
	}

	parsed, err := nu.Parse(url)
	if err != nil {
		return false
	}

	if len(parsed.Scheme) == 0 {
		return false
	}

	return true
}

func cleanURL(url string) string {
	if !strings.HasSuffix(url, "/") {
		return fmt.Sprintf("%s/", url)
	}
	return url
}

func getURL(url, path string) string {
	u := cleanURL(url)
	return fmt.Sprintf("%s%s", u, path)
}
