# WASI
Golang(tinygo) web assembly interface types. 

Currently, we are focused on implementing bindings when needed for `WASI` using wasip2 and specifically the component model. 

Tinygo support for wasip2 is new. We leverage the dev branch and build locally to take advantage of latest development addtions. 

## HTTP 

A custom roundtrip transport exists that can be used for http clients. 

```go
//go:build !wasip1

package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/hayride-dev/wasi/http/transport"
)

func main() {
	client := &http.Client{
		Transport: transport.NewWasiRoundTripper(), 
	}
	req, err := http.NewRequest("GET", "https://postman-echo.com/get", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Set("User-Agent", "wasi-http-client")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("response:", string(body))
}
```

### Running with a WASI runtime 

Our prefered runtime is [wasmtime](https://github.com/bytecodealliance/wasmtime). Other WASI enabled runtimes should be capatible, but use at your own risk! 

Running the above example:
- Building tinygo: `tinygo build -tags purego -o main.wasm -target=wasip2 -wit-package ../wit/wit --wit-world platform `. 
- Run with wasmtime-cli: `WASMTIME_LOG=wasmtime_wasi=trace WASMTIME_BACKTRACE_DETAILS=1 wasmtime -S http --wasm component-model main.wasm`.

