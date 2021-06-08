package gbrotli

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/gin-gonic/gin"
)

type brotliWriter struct {
	gin.ResponseWriter
	writer *brotli.Writer
}

func (br *brotliWriter) WriteString(s string) (int, error) {
	return br.writer.Write([]byte(s))
}

func (br *brotliWriter) Write(data []byte) (int, error) {
	return br.writer.Write(data)
}

// Fix: https://github.com/mholt/caddy/issues/38
func (br *brotliWriter) WriteHeader(code int) {
	br.Header().Del("Content-Length")
	br.ResponseWriter.WriteHeader(code)
}

var (
	// DefaultCompression Quality: 4 LGWin: 11
	// from 0-11. 4 will bring the files faster than 11. the higher quality the slower compression time
	DefaultCompression = Options{
		WriterOptions: brotli.WriterOptions{
			Quality: 4,
			LGWin:   11,
		},
		SkipExtensions: []string{".png", ".gif", ".jpeg", ".jpg", ".mp3", ".mp4"},
	}
)

// Options is a wrapper for cbrotli.WriterOptions
type Options struct {
	brotli.WriterOptions
	SkipExtensions []string
}

// Brotli is a middleware function
func Brotli(options Options) gin.HandlerFunc {

	return func(c *gin.Context) {
		if !shouldCompress(c.Request, &options) {
			return
		}

		brWriter := brotli.NewWriterOptions(c.Writer, brotli.WriterOptions{
			Quality: options.Quality,
			LGWin:   options.LGWin,
		})

		c.Header("Content-Encoding", "br")
		c.Header("Vary", "Accept-Encoding")
		c.Writer = &brotliWriter{c.Writer, brWriter}

		defer func() {
			brWriter.Close()
			c.Header("Content-Length", fmt.Sprint(c.Writer.Size()))
		}()
		c.Next()
	}
}

func containsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func shouldCompress(req *http.Request, options *Options) bool {
	if !strings.Contains(req.Header.Get("Accept-Encoding"), "br") ||
		strings.Contains(req.Header.Get("Connection"), "Upgrade") ||
		strings.Contains(req.Header.Get("Content-Type"), "text/event-stream") {

		return false
	}

	extension := filepath.Ext(req.URL.Path)
	if len(extension) < 4 { // fast path
		return true
	}

	if skip := containsString(options.SkipExtensions, extension); skip {
		return false
	} else {
		return true
	}
}
