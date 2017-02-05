package testhelpers

import "net/http"

type TestingHTTPTransport struct {
	OnRequest         func(req *http.Request) (*http.Response, error)
	originalTransport http.RoundTripper
}

func FakeTransport(f func(req *http.Request) (*http.Response, error)) TestingHTTPTransport {
	transport := TestingHTTPTransport{
		OnRequest:         f,
		originalTransport: http.DefaultClient.Transport,
	}

	http.DefaultClient.Transport = transport
	return transport
}

func (transport TestingHTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return transport.OnRequest(req)
}

func (transport *TestingHTTPTransport) Close() {
	http.DefaultClient.Transport = transport.originalTransport
}
