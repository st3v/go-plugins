package awsxray

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/micro/go-micro/errors"
	"github.com/micro/go-micro/metadata"
	"golang.org/x/net/context"
)

// getHTTP returns a http struct
func getHTTP(url, method string, err error) http {
	return http{
		Request: request{
			Method: method,
			URL:    url,
		},
		Response: response{
			Status: getStatus(err),
		},
	}
}

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

// getStatus returns a status code from the error
func getStatus(err error) int {
	// no error
	if err == nil {
		return 200
	}

	// try get errors.Error
	if e, ok := err.(*errors.Error); ok {
		return int(e.Code)
	}

	// try parse marshalled error
	if e := errors.Parse(err.Error()); e.Code > 0 {
		return int(e.Code)
	}

	// could not parse, 500
	return 500
}

// getTraceId returns trace header or generates a new one
func getTraceId(md metadata.Metadata) string {
	fn := func(header string) string {
		for _, h := range strings.Split(header, ";") {
			th := strings.TrimSpace(h)
			if strings.HasPrefix(th, "Root=") {
				return strings.TrimPrefix(th, "Root=")
			}
		}

		// return as is
		return header
	}

	// try as is
	if h, ok := md[TraceHeader]; ok {
		return fn(h)
	}

	// try lower case
	if h, ok := md[strings.ToLower(TraceHeader)]; ok {
		return fn(h)
	}

	// generate new one, probably a bad idea...
	return fmt.Sprintf("%d-%x-%x", 1, time.Now().Unix(), getRandom(12))
}

// getParentId returns parent header or blank
func getParentId(md metadata.Metadata) string {
	fn := func(header string) string {
		for _, h := range strings.Split(header, ";") {
			th := strings.TrimSpace(h)
			if strings.HasPrefix(th, "Parent=") {
				return strings.TrimPrefix(th, "Parent=")
			}
		}

		// return nothing
		return ""
	}

	// try as is
	if h, ok := md[TraceHeader]; ok {
		return fn(h)
	}

	// try lower case
	if h, ok := md[strings.ToLower(TraceHeader)]; ok {
		return fn(h)
	}

	return ""
}

func setTraceId(header, traceId string) string {
	headers := strings.Split(header, ";")
	traceHeader := fmt.Sprintf("Root=%s", traceId)

	for i, h := range headers {
		th := strings.TrimSpace(h)
		// get Root=Id match
		if strings.HasPrefix(th, "Root=") {
			// set trace header
			headers[i] = traceHeader
			// return entire header
			return strings.Join(headers, "; ")
		}
	}

	// no match; set new trace header as first entry
	return strings.Join(append([]string{traceHeader}, headers...), "; ")
}

func setParentId(header, parentId string) string {
	headers := strings.Split(header, ";")
	parentHeader := fmt.Sprintf("Parent=%s", parentId)

	for i, h := range headers {
		th := strings.TrimSpace(h)
		// get Parent=Id match
		if strings.HasPrefix(th, "Parent=") {
			// set parent header
			headers[i] = parentHeader
			// return entire header
			return strings.Join(headers, "; ")
		}
	}

	// no match; set new parent header
	return strings.Join(append(headers, parentHeader), "; ")
}

func newContext(ctx context.Context, s *segment) context.Context {
	md, _ := metadata.FromContext(ctx)

	// make copy to avoid races
	newMd := metadata.Metadata{}
	for k, v := range md {
		newMd[k] = v
	}

	// set trace id in header
	newMd[TraceHeader] = setTraceId(newMd[TraceHeader], s.TraceId)
	// set parent id in header
	newMd[TraceHeader] = setParentId(newMd[TraceHeader], s.ParentId)
	// store segment in context
	ctx = context.WithValue(ctx, contextSegmentKey{}, s)
	// store metadata in context
	ctx = metadata.NewContext(ctx, newMd)

	return ctx
}
