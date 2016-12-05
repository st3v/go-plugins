package awsxray

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/xray"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/metadata"
	"github.com/micro/go-micro/server"
	"golang.org/x/net/context"
)

type xrayWrapper struct {
	x *xray.XRay
	client.Client
}

type segment struct {
	Name      string `json:"name"`
	Id        string `json:"id"`
	TraceId   string `json:"trace_id"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
	Type      string `json:"type,omitempty"`
	ParentId  string `json:"parent_id,omitempty"`
}

var (
	ParentHeader = "X-Amzn-Parent-Id"
	TraceHeader  = "X-Amzn-Trace-Id"
)

// getRandom generates a random byte slice
func getRandom(i int) []byte {
	b := make([]byte, i)
	for {
		// keep trying till we get it
		if _, err := rand.Read(b); err != nil {
			continue
		}
		return b
	}
}

// getTraceId returns trace header or generates a new one
func getTraceId(md metadata.Metadata) string {
	// try as is
	if h, ok := md[TraceHeader]; ok {
		return h
	}

	// try lower case
	if h, ok := md[strings.ToLower(TraceHeader)]; ok {
		return h
	}

	// generate new one, probably a bad idea...
	return fmt.Sprintf("%d-%x-%x", 1, time.Now().Unix(), getRandom(12))
}

// getParentId returns parent header or blank
func getParentId(md metadata.Metadata) string {
	// try as is
	if h, ok := md[ParentHeader]; ok {
		return h
	}

	// try lower case
	if h, ok := md[strings.ToLower(ParentHeader)]; ok {
		return h
	}

	return ""
}

// record sends the trace segment
func record(x *xray.XRay, s *segment) {
	b, _ := json.Marshal(s)

	// ignoring response and error
	x.PutTraceSegments(&xray.PutTraceSegmentsInput{
		TraceSegmentDocuments: []*string{
			aws.String("TraceSegmentDocument"),
			aws.String(string(b)),
		},
	})
}

func (x *xrayWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	md, _ := metadata.FromContext(ctx)
	start := time.Now()

	defer func() {
		record(x.x, &segment{
			Id:        fmt.Sprintf("%x", getRandom(8)),
			Name:      fmt.Sprintf("%s.%s", req.Service(), req.Method()),
			TraceId:   getTraceId(md),
			StartTime: start.UnixNano() / 1e6,
			EndTime:   time.Now().UnixNano() / 1e6,
			Type:      "subsegment", // eh?
			ParentId:  getParentId(md),
		})
	}()

	return x.Client.Call(ctx, req, rsp, opts...)
}

// NewCallWrapper accepts xray.XRay and returns a Trace Call Wrapper for individual node calls made by the client
func NewCallWrapper(x *xray.XRay) client.CallWrapper {
	return func(cf client.CallFunc) client.CallFunc {
		return func(ctx context.Context, addr string, req client.Request, rsp interface{}, opts client.CallOptions) error {
			md, _ := metadata.FromContext(ctx)
			start := time.Now()

			defer func() {
				record(x, &segment{
					Id:        fmt.Sprintf("%x", getRandom(8)),
					Name:      addr,
					TraceId:   getTraceId(md),
					StartTime: start.UnixNano() / 1e6,
					EndTime:   time.Now().UnixNano() / 1e6,
					Type:      "subsegment", // eh?
					ParentId:  getParentId(md),
				})
			}()

			return cf(ctx, addr, req, rsp, opts)
		}
	}
}

// NewClientWrapper accepts xray.XRay and returns a Trace Client Wrapper which tracks high level service calls
func NewClientWrapper(x *xray.XRay) client.Wrapper {
	return func(c client.Client) client.Client {
		return &xrayWrapper{x, c}
	}
}

// NewHandlerWrapper accepts xray.XRay and returns a Trace Handler Wrapper
func NewHandlerWrapper(x *xray.XRay) server.HandlerWrapper {
	return func(h server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			// Set our own parent id to be used be further calls
			parentId := fmt.Sprintf("%x", getRandom(8))
			md, _ := metadata.FromContext(ctx)
			kmd := metadata.Metadata{}
			for k, v := range md {
				kmd[k] = v
			}
			kmd[ParentHeader] = parentId
			start := time.Now()

			defer func() {
				record(x, &segment{
					Id:        fmt.Sprintf("%x", getRandom(8)),
					Name:      fmt.Sprintf("%s.%s", req.Service(), req.Method()),
					TraceId:   getTraceId(md),
					StartTime: start.UnixNano() / 1e6,
					EndTime:   time.Now().UnixNano() / 1e6,
				})
			}()

			ctx = metadata.NewContext(ctx, kmd)
			return h(ctx, req, rsp)
		}
	}
}
