package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/makimaki04/go-metrics-agent.git/internal/compress"
)

// GzipMiddleware - middleware for compressing the response
// decompress the request body if it is gzip encoded
// compress the response body if request header contains content-encoding: gzip
func GzipMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportGzip := strings.Contains(acceptEncoding, "gzip")
		if supportGzip {
			cw := compress.NewCompressWriter(w)
			ow = cw
			defer compress.ReleaseWriter(cw.Writer)
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			gr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Invalid gzip body", http.StatusBadRequest)
				return
			}
			r.Body = gr
			defer gr.Close()
		}

		h.ServeHTTP(ow, r)
	}
}
