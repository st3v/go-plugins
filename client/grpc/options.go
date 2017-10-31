// Package grpc provides a gRPC options
package grpc

import (
	"github.com/micro/go-micro/client"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type codecsKey struct{}

// gRPC Codec to be used to encode/decode requests for a given content type
func Codec(contentType string, c grpc.Codec) client.Option {
	return func(o *client.Options) {
		codecs := make(map[string]grpc.Codec)
		if o.Context == nil {
			o.Context = context.Background()
		}
		if v := o.Context.Value(codecsKey{}); v != nil {
			codecs = v.(map[string]grpc.Codec)
		}
		codecs[contentType] = c
		o.Context = context.WithValue(o.Context, codecsKey{}, codecs)
	}
}
