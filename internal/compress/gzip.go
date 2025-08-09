package compress

import (
	"compress/gzip"
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
	return c.Writer.Write(b)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.ResponseWriter.Header().Set("Content-Encoding", "gzip")
	}
	c.ResponseWriter.WriteHeader(statusCode)
}