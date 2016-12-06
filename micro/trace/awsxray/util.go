package awsxray

import (
	"crypto/rand"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// getIp naively returns an ip for the request
func getIp(r *http.Request) string {
	for _, h := range []string{"X-Forwarded-For", "X-Real-Ip"} {
		for _, ip := range strings.Split(r.Header.Get(h), ",") {
			if len(ip) == 0 {
				continue
			}
			realIP := net.ParseIP(strings.Replace(ip, " ", "", -1))
			return realIP.String()
		}
	}

	// not found in header
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// just return remote addr
		return r.RemoteAddr
	}

	return host
}

// getRandom generates a random byte slice
func getRandom(i int) string {
	b := make([]byte, i)
	for {
		// keep trying till we get it
		if _, err := rand.Read(b); err != nil {
			continue
		}
		return fmt.Sprintf("%x", b)
	}
}

// getTraceId returns trace header or generates a new one
func getTraceId(hdr http.Header) string {
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
	if h := hdr.Get(TraceHeader); len(h) > 0 {
		return fn(h)
	}

	// generate new one, probably a bad idea...
	return fmt.Sprintf("%d-%x-%s", 1, time.Now().Unix(), getRandom(12))
}

// getParentId returns parent header or blank
func getParentId(hdr http.Header) string {
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
	if h := hdr.Get(TraceHeader); len(h) > 0 {
		return fn(h)
	}

	return ""
}

func setTraceId(header, traceId string) string {
	traceHeader := fmt.Sprintf("Root=%s", traceId)

	// no existing header?
	if len(header) == 0 {
		return traceHeader
	}

	headers := strings.Split(header, ";")

	for i, h := range headers {
		th := strings.TrimSpace(h)

		if len(th) == 0 {
			continue
		}

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
	parentHeader := fmt.Sprintf("Parent=%s", parentId)

	// no existing header?
	if len(header) == 0 {
		return parentHeader
	}

	headers := strings.Split(header, ";")

	for i, h := range headers {
		th := strings.TrimSpace(h)

		if len(th) == 0 {
			continue
		}

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
