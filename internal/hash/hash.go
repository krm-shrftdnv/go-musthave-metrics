package hash

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	customHttp "github.com/krm-shrftdnv/go-musthave-metrics/internal/http"
)

type hashWriter struct {
	hashKey string
	w       http.ResponseWriter
}

func newHashWriter(hashKey string, w http.ResponseWriter) *hashWriter {
	return &hashWriter{
		hashKey: hashKey,
		w:       w,
	}
}

func (h *hashWriter) Write(p []byte) (int, error) {
	if h.hashKey != "" {
		hash, err := hash(p, h.hashKey)
		if err != nil {
			return 0, err
		}
		h.w.Header().Set("HashSHA256", hash)
	}
	return h.w.Write(p)
}

func (h *hashWriter) Header() http.Header {
	return h.w.Header()
}

func (h *hashWriter) WriteHeader(statusCode int) {
	h.w.WriteHeader(statusCode)
}

func New(key string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hashFunc := func(w http.ResponseWriter, r *http.Request) {
			ow := w
			if key != "" {
				acceptedHash := r.Header.Get("HashSHA256")
				if acceptedHash != "" {
					body, err := io.ReadAll(r.Body)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					err = r.Body.Close()
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					hash, err := hash(body, key)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					if !hmac.Equal([]byte(hash), []byte(acceptedHash)) {
						w.WriteHeader(http.StatusBadRequest)
						return
					}
				}
				hw := newHashWriter(key, w)
				ow = hw
			}
			next.ServeHTTP(ow, r)
		}
		return http.HandlerFunc(hashFunc)
	}
}

func HashRequest(key string) customHttp.Middleware {
	return func(rt http.RoundTripper) http.RoundTripper {
		return customHttp.InternalRoundTripper(func(req *http.Request) (*http.Response, error) {
			if key != "" {
				body, err := io.ReadAll(req.Body)
				if err != nil {
					return nil, err
				}
				err = req.Body.Close()
				if err != nil {
					return nil, err
				}
				hash, err := hash(body, key)
				if err != nil {
					return nil, err
				}
				header := req.Header
				if header == nil {
					header = make(http.Header)
				}
				header.Set("HashSHA256", hash)
				req.Header = header
				req.Body = io.NopCloser(bytes.NewReader(body))
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
	return hex.EncodeToString(dst), nil
}
