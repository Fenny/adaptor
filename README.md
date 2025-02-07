# Adaptor

![Release](https://img.shields.io/github/release/gofiber/adaptor.svg)
[![Discord](https://img.shields.io/badge/discord-join%20channel-7289DA)](https://gofiber.io/discord)
![Test](https://github.com/gofiber/adaptor/workflows/Test/badge.svg)
![Security](https://github.com/gofiber/adaptor/workflows/Security/badge.svg)
![Linter](https://github.com/gofiber/adaptor/workflows/Linter/badge.svg)

Converter for net/http handlers to/from Fiber request handlers, special thanks to [@arsmn](https://github.com/arsmn)!

### Install
```
go get -u github.com/gofiber/fiber
go get -u github.com/gofiber/adaptor
```

### Functions
| Name | Signature | Description
| :--- | :--- | :---
| HTTPHandler | `HTTPHandler(h http.Handler) func(*fiber.Ctx)` | http.Handler -> Fiber handler
| HTTPHandlerFunc | `HTTPHandlerFunc(h http.HandlerFunc) func(*fiber.Ctx)` | http.HandlerFunc -> Fiber handler
| FiberHandler | `FiberHandler(h func(*fiber.Ctx)) http.Handler` | Fiber handler -> http.Handler
| FiberHandlerFunc | `FiberHandlerFunc(h func(*fiber.Ctx)) http.HandlerFunc` | Fiber handler -> http.HandlerFunc

### net/http to Fiber
```go
package main

import (
	"fmt"
	"net/http"

	"github.com/gofiber/adaptor"
	"github.com/gofiber/fiber"
)

func main() {
	// New fiber app
	app := fiber.New()

	// http.Handler -> func(*fiber.Ctx)
	app.Get("/", adaptor.HTTPHandler(handler(greet)))

	// http.HandlerFunc -> func(*fiber.Ctx)
	app.Get("/func", adaptor.HTTPHandlerFunc(greet))

	// Listen on port 3000
	app.Listen(3000)
}

func handler(f http.HandlerFunc) http.Handler {
	return http.HandlerFunc(f)
}

func greet(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello World!")
}
```

### Fiber to net/http
```go
package main

import (
	"net/http"

	"github.com/gofiber/adaptor"
	"github.com/gofiber/fiber"
)

func main() {
	// func(c *fiber.Ctx) -> http.Handler
	http.Handle("/", adaptor.FiberHandler(greet))

  	// func(c *fiber.Ctx) -> http.HandlerFunc
	http.HandleFunc("/func", adaptor.FiberHandlerFunc(greet))

	// Listen on port 3000
	http.ListenAndServe(":3000", nil)
}

func greet(c *fiber.Ctx) {
	c.Send("Hello World!")
}
```
