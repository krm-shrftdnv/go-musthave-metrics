package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"io"
	"net/http"

	customHttp "github.com/krm-shrftdnv/go-musthave-metrics/internal/http"
)

type hashWriter struct {
	hashKey string
	w       http.ResponseWriter
}

type hashReader struct {
	hashKey string
	r       io.ReadCloser
}

func New(key string) func(next http.HandlerFunc) http.HandlerFunc {
}

func HashRequest(key string) customHttp.Middleware {
	return func(rt http.RoundTripper) http.RoundTripper {
		return customHttp.InternalRoundTripper(func(req *http.Request) (*http.Response, error) {
			if key != "" {
				body := req.Body
				var bodyBytes []byte
				_, err := body.Read(bodyBytes)
				if err != nil {
					return nil, err
				}
				hash, err := hash(bodyBytes, key)
				if err != nil {
					return nil, err
				}
				header := req.Header
				if header == nil {
					header = make(http.Header)
				}
				header.Set("HashSHA256", hash)
			}
			return rt.RoundTrip(req)
		})
	}
}

func hash(body []byte, hashKey string) (string, error) {
	h := hmac.New(sha256.New, []byte(hashKey))
	_, err := h.Write(body)
	if err != nil {
		return "", err
	}
	dst := h.Sum(nil)
	return string(dst), nil
}
