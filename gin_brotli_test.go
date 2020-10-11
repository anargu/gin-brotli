package gin_brotli

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strconv"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var (
	testResponse = gin.H{
		"message": "a simple message",
	}
	testJSONResponse    = "{\"message\":\"a simple message\"}"
	testReverseResponse = "{\"message\":\"a simple message\"}"
)

type closeNotifyingRecorder struct {
	*httptest.ResponseRecorder
	closed chan bool
}

func newCloseNotifyingRecorder() *closeNotifyingRecorder {
	return &closeNotifyingRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
}

func (c *closeNotifyingRecorder) close() {
	c.closed <- true
}

func (c *closeNotifyingRecorder) CloseNotify() <-chan bool {
	return c.closed
}

type rServer struct{}

func (s *rServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprint(rw, testReverseResponse)
}

func newServerWithProxy() *gin.Engine {

	// init reverse proxy server
	rServer := httptest.NewServer(new(rServer))
	target, _ := url.Parse(rServer.URL)
	rp := httputil.NewSingleHostReverseProxy(target)

	router := gin.New()
	router.Use(Brotli(DefaultCompression))
	router.GET("/", func(c *gin.Context) {

		c.Header("Content-Length", strconv.Itoa(len(testJSONResponse)))
		c.JSON(http.StatusOK, testResponse)
	})
	router.Any("/reverse", func(c *gin.Context) {
		rp.ServeHTTP(c.Writer, c.Request)
	})
	return router
}

func newBenchmarkServer() *gin.Engine {
	// init reverse proxy server
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(Brotli(DefaultCompression))
	router.GET("/", func(c *gin.Context) {

		data, err := ioutil.ReadFile("testdata/data.json")
		if err != nil {
			panic(err)
		}
		c.Header("Content-Length", strconv.Itoa(len(data)))
		c.JSON(http.StatusOK, data)
	})
	return router
}

func TestBrotli(t *testing.T) {

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept-Encoding", "br")

	w := httptest.NewRecorder()
	r := newServerWithProxy()
	r.ServeHTTP(w, req)

	assert.Equal(t, w.Code, 200)
	assert.Equal(t, w.Header().Get("Content-Encoding"), "br")
	assert.Equal(t, w.Header().Get("Vary"), "Accept-Encoding")
	assert.NotEqual(t, w.Header().Get("Content-Length"), "0")
	assert.NotEqual(t, len(testResponse), w.Body.Len())
	assert.Equal(t, fmt.Sprint(w.Body.Len()), w.Header().Get("Content-Length"))

	// fmt.Printf("\n+++Test w.Body.Len()+++\n%v\n+++\n", w.Body.Len())
	// fmt.Printf("\n+++Test w.Header().Get(\"Content-Length\")+++\n%s\n+++\n", w.Header().Get("Content-Length"))

	br := brotli.NewReader(w.Body)
	body, _ := ioutil.ReadAll(br)

	// fmt.Printf("\n+++Test string(body)+++\n%s\n+++\n", string(body))
	// fmt.Printf("\n+++Test testJSONResponse+++\n%s\n+++\n", testJSONResponse)

	assert.Equal(t, string(body), testJSONResponse)
}

func BenchmarkBrotli(b *testing.B) {
	b.StopTimer()

	r := newBenchmarkServer()

	for i := 0; i < b.N; i++ {
		req, err := http.NewRequest(http.MethodGet, "/json", nil)
		if err != nil {
			b.Fatalf("could not create request %v", err)
		}
		req.Header.Add("Accept-Encoding", "br")

		rec := httptest.NewRecorder()
		b.StartTimer()
		r.ServeHTTP(rec, req)
		b.StopTimer()
	}
}

func TestNotSupportBrotli(t *testing.T) {

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept-Encoding", "gzip")

	w := httptest.NewRecorder()
	r := newServerWithProxy()
	r.ServeHTTP(w, req)

	assert.Equal(t, w.Code, 200)
	assert.NotEqual(t, w.Header().Get("Content-Encoding"), "br")
	assert.NotEqual(t, w.Header().Get("Vary"), "Accept-Encoding")
	assert.Equal(t, w.Header().Get("Vary"), "")
	assert.NotEqual(t, w.Header().Get("Content-Length"), "0")
	assert.Equal(t, len(testJSONResponse), w.Body.Len())
	assert.Equal(t, fmt.Sprint(w.Body.Len()), w.Header().Get("Content-Length"))

	// fmt.Printf("\n+++Test w.Body.Len()+++\n%v\n+++\n", w.Body.Len())
	// fmt.Printf("\n+++Test w.Header().Get(\"Content-Length\")+++\n%s\n+++\n", w.Header().Get("Content-Length"))

	// fmt.Printf("\n+++Test string(body)+++\n%s\n+++\n", string(body))
	// fmt.Printf("\n+++Test testJSONResponse+++\n%s\n+++\n", testJSONResponse)

	assert.Equal(t, w.Body.String(), testJSONResponse)
}

func TestBrotliWithReverseProxy(t *testing.T) {

	req, _ := http.NewRequest("GET", "/reverse", nil)
	req.Header.Add("Accept-Encoding", "br")

	w := newCloseNotifyingRecorder()
	r := newServerWithProxy()
	r.ServeHTTP(w, req)

	assert.Equal(t, w.Code, 200)
	assert.Equal(t, w.Header().Get("Content-Encoding"), "br")
	assert.Equal(t, w.Header().Get("Vary"), "Accept-Encoding")
	assert.NotEqual(t, w.Header().Get("Content-Length"), "0")
	assert.Equal(t, fmt.Sprint(w.Body.Len()), w.Header().Get("Content-Length"))

	fmt.Printf("\n+++Test w.Body.Len()+++\n%v\n+++\n", w.Body.Len())
	fmt.Printf("\n+++Test w.Header().Get(\"Content-Length\")+++\n%s\n+++\n", w.Header().Get("Content-Length"))

	br := brotli.NewReader(w.Body)
	body, _ := ioutil.ReadAll(br)

	fmt.Printf("\n+++Test string(body)+++\n%v\n+++\n", string(body))
	fmt.Printf("\n+++Test testJSONResponse+++\n%v\n+++\n", testReverseResponse)
	assert.Equal(t, string(body), testReverseResponse)
}
