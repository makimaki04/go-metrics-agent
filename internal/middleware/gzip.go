package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/makimaki04/go-metrics-agent.git/internal/compress"
)

func GzipMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportGzip := strings.Contains(acceptEncoding, "gzip")
		if supportGzip {
			cw := compress.NewCompressWriter(w)
			ow = cw
			defer cw.Writer.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")

		if sendsGzip {
			gr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Invalid gzip body", http.StatusBadRequest)
				return
			}
			defer gr.Close()
			r.Body = gr
		}

		h.ServeHTTP(ow, r)
	}
}