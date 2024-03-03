# bytes pool

Simple `[]byte` buffer pool for golang backend by `sync.Pool`.

[![Go Reference](https://pkg.go.dev/badge/github.com/IrineSistiana/bytespool.svg)](https://pkg.go.dev/github.com/IrineSistiana/bytespool)

```go
package main

import "github.com/IrineSistiana/bytespool"

func main() {
	b := Get(1024)
	Release(b)
}
```
