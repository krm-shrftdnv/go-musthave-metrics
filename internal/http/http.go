package http

import "net/http"

type InternalRoundTripper func(*http.Request) (*http.Response, error)

func (rt InternalRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt(req)
}

type Middleware func(http.RoundTripper) http.RoundTripper

func Chain(rt http.RoundTripper, middlewares ...Middleware) http.RoundTripper {
	if rt == nil {
		rt = http.DefaultTransport
	}

	for _, m := range middlewares {
		rt = m(rt)
	}

	return rt
}
