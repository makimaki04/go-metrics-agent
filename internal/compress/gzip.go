package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"sync"
)

type CompressWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

var gzipWriterPool = sync.Pool{
	New: func() any {
		return new(gzip.Writer)
	},
}

func AcquireWriter(w io.Writer) *gzip.Writer {
	gw := gzipWriterPool.Get().(*gzip.Writer)
	gw.Reset(w)

	return gw
}

func ReleaseWriter(gw *gzip.Writer) {
	gw.Close()
	gw.Reset(io.Discard)

	gzipWriterPool.Put(gw)
}

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
