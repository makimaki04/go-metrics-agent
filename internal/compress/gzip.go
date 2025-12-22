package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"sync"
)

// CompressWriter - wrapper for http.ResponseWriter that compresses response using gzip
// ResponseWriter - underlying http.ResponseWriter
// Writer - gzip writer for compression
type CompressWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

var gzipWriterPool = sync.Pool{
	New: func() any {
		return new(gzip.Writer)
	},
}

// AcquireWriter - acquires a gzip.Writer from the pool and resets it for use
// w - writer to reset the gzip writer for
// returns a gzip.Writer ready for use
func AcquireWriter(w io.Writer) *gzip.Writer {
	gw := gzipWriterPool.Get().(*gzip.Writer)
	gw.Reset(w)

	return gw
}

// ReleaseWriter - releases a gzip.Writer back to the pool
// closes the writer and resets it before returning to pool
// gw - gzip writer to release
func ReleaseWriter(gw *gzip.Writer) {
	gw.Close()
	gw.Reset(io.Discard)

	gzipWriterPool.Put(gw)
}

// NewCompressWriter - creates a new CompressWriter
// wraps the provided ResponseWriter with gzip compression
// w - http.ResponseWriter to wrap
// returns a new CompressWriter instance
func NewCompressWriter(w http.ResponseWriter) *CompressWriter {
	return &CompressWriter{
		ResponseWriter: w,
		Writer:         AcquireWriter(w),
	}
}

func (c *CompressWriter) Write(b []byte) (int, error) {
	return c.Writer.Write(b)
}

func (c *CompressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.ResponseWriter.Header().Set("Content-Encoding", "gzip")
	}
	c.ResponseWriter.WriteHeader(statusCode)
}
