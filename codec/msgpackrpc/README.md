# MsgPack RPC Codec

## Usage

```go
package main

import (
    "github.com/micro/go-plugins/codec/msgpackrpc"
    "github.com/micro/go-micro"
    "github.com/micro/go-micro/server"
)

func main() {
    server := server.NewServer(
        server.Codec("application/msgpack", msgpackrpc.NewCodec),
    )

    service := micro.NewService(
        micro.Server(server),
    )

	// ...
}
```
