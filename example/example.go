package main

import (
	"io/ioutil"
	"net/http"

	br "github.com/anargu/gin_brotli"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Use(br.Brotli(br.DefaultCompression))
	r.GET("/json", func(c *gin.Context) {
		bytes, err := ioutil.ReadFile("../testdata/data.json")
		if err != nil {
			panic(err)
		}

		c.JSON(http.StatusOK, bytes)
	})
	r.Run()
}
