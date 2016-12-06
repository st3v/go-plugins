package awsxray

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/micro/go-micro/errors"
	"github.com/micro/go-micro/metadata"
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
			if strings.HasPrefix(h, "Root=") {
				return strings.TrimPrefix(h, "Root=")
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
			if strings.HasPrefix(h, "Parent=") {
				return strings.TrimPrefix(h, "Parent=")
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
