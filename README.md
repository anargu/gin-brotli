# Brotli gin's middleware

Gin middleware to enable [Brotli](https://github.com/google/brotli) support.

NOTE: this repo is an adaptation of how gzip middleware is implemented. I'll try to add new features.

## Requirements

~~Install Brotli, [see here](https://github.com/google/brotli).~~

~~Install brotli package for go (cbrotli). Copy [github.com/google/brotli/tree/master/go/cbrotli](github.com/google/brotli/tree/master/go/cbrotli) package into GOPATH/ directory~~

#### [Update] gin-brotli does not depend on cbrotli installed. Now it uses brotli from [andybalholm/brotli](https://github.com/andybalholm/brotli)

## Install

    go get github.com/anargu/gin-brotli

## How to use

    package main

    import (
        "fmt"
        "time"
	    "net/http"

        brotli "github.com/anargu/gin-brotli"
        "github.com/gin-gonic/gin"
    )

    func main() {
        r := gin.Default()
        r.Use(brotli.Brotli(brotli.DefaultCompression))
        r.GET("/hello", func(c *gin.Context) {
            c.String(http.StatusOK, fmt.Sprintf("World at %s", time.Now()))
        })

        // Listen and Server in 0.0.0.0:8080
        r.Run(":8080")
    }

## Test it

	cd example/

    go run example.go
    
In Another terminal

    curl -X GET http://localhost:8080/json

## TODO

- Add *fallback* feature: If brotli is not supported in browser then the request will be handled by gzip compression. And if it's not supported by the browser yet, the request is going to be send as is (without compression).

