package compress

import (
	"compress/gzip"
	"io"
	"net/http"

	"github.com/makimaki04/go-metrics-agent.git/internal/pool"
)

// CompressWriter - wrapper for http.ResponseWriter that compresses response using gzip
// ResponseWriter - underlying http.ResponseWriter
// Writer - gzip writer for compression
type CompressWriter struct {
	http.ResponseWriter
	Wrapper *GzipWriterWrapper
}

type GzipWriterWrapper struct {
	writer *gzip.Writer
}

func (g *GzipWriterWrapper) Reset() {
	if g.writer != nil {
		g.writer.Close()
		g.writer.Reset(io.Discard)
	}
}

func (g *GzipWriterWrapper) ResetTo(w io.Writer) {
	if g.writer != nil {
		g.writer.Reset(w)
	}
}

func (g *GzipWriterWrapper) Writer() *gzip.Writer {
	return g.writer
}

var gzipWriterPool = pool.New[*GzipWriterWrapper](func() *GzipWriterWrapper {
	return &GzipWriterWrapper{writer: new(gzip.Writer)}
})

// AcquireWriter - acquires a gzip.Writer from the pool and resets it for use
// w - writer to reset the gzip writer for
// returns a gzip.Writer ready for use
func AcquireWriter(w io.Writer) *GzipWriterWrapper {
	gw := gzipWriterPool.Get()
	gw.ResetTo(w)

	return gw
}

// ReleaseWriter - releases a gzip.Writer back to the pool
// closes the writer and resets it before returning to pool
// gw - gzip writer to release
func ReleaseWriter(cw *CompressWriter) {
	gzipWriterPool.Put(cw.Wrapper)
}

// NewCompressWriter - creates a new CompressWriter
// wraps the provided ResponseWriter with gzip compression
// w - http.ResponseWriter to wrap
// returns a new CompressWriter instance
func NewCompressWriter(w http.ResponseWriter) *CompressWriter {
	return &CompressWriter{
		ResponseWriter: w,
		Wrapper:        AcquireWriter(w),
	}
}

func (c *CompressWriter) Write(b []byte) (int, error) {
	return c.Wrapper.Writer().Write(b)
}

func (c *CompressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.ResponseWriter.Header().Set("Content-Encoding", "gzip")
	}
	c.ResponseWriter.WriteHeader(statusCode)
}
