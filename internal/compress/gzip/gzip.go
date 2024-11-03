package gzip

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	customHttp "github.com/krm-shrftdnv/go-musthave-metrics/internal/http"
)

type compressWriter struct {
	w            http.ResponseWriter
	zw           *gzip.Writer
	responseData struct {
		status           int
		compressibleType bool
	}
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	contentType := c.w.Header().Get("Content-Type")
	c.responseData.compressibleType = strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "text/html")
	if c.responseData.status < 300 && c.responseData.compressibleType {
		c.w.Header().Set("Content-Encoding", "gzip")
		return c.zw.Write(p)
	}
	return c.w.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	contentType := c.w.Header().Get("Content-Type")
	c.responseData.compressibleType = strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "text/html")
	if c.responseData.status < 300 && c.responseData.compressibleType {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
	c.responseData.status = statusCode
}

func (c *compressWriter) Close() error {
	return c.zw.Close()
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c *compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func CompressRequestBody(h http.Handler) http.Handler {
	gzipFn := func(w http.ResponseWriter, r *http.Request) {
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		ow := w
		if supportsGzip {
			cw := newCompressWriter(w)
			ow = cw
			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}
		h.ServeHTTP(ow, r)
	}
	return http.HandlerFunc(gzipFn)
}

func CompressRequest() customHttp.Middleware {
	return func(rt http.RoundTripper) http.RoundTripper {
		return customHttp.InternalRoundTripper(func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			req.Body.Close()
			if err != nil {
				return nil, err
			}
			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)
			if _, err := gz.Write(body); err != nil {
				return nil, err
			}
			if err := gz.Close(); err != nil {
				return nil, err
			}
			header := req.Header
			if header == nil {
				header = make(http.Header)
			}
			req.Body = io.NopCloser(&buf)
			req.ContentLength = int64(buf.Len())
			header.Set("Content-Encoding", "gzip")
			header.Set("Accept-Encoding", "gzip")
			return rt.RoundTrip(req)
		})
	}
}
