# Brotli gin's middleware

Gin middleware to enable [Brotli](https://github.com/google/brotli) support.

NOTE: this repo is an adaptation of how gzip middleware is implemented. except for sync.Pool that will going to be implemented soon, also will add new features.

## Requirements

Install Brotli, [see here](https://github.com/google/brotli).

## Install

    TODO

## How to use

    TODO

## TODO

- Add like a *fallback*: If brotli is not supported in browser then the request will be handled by gzip compression. And if it's not supported by the browser yet, the request is going to be send as is (without compression).
