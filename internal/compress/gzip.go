package compress

import (
	"compress/gzip"
	"io"
	"net/http"
)

type compressWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

func NewCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		ResponseWriter: w,
		Writer: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Write(b []byte) (int, error) {
	header := c.Header().Get("Content-Type")
	if header == "application/json" || header == "text/html" {
		return c.Writer.Write(b)
	}
	return c.ResponseWriter.Write(b)
}

type compressReader struct {
	io.ReadCloser
	reader *gzip.Reader
}