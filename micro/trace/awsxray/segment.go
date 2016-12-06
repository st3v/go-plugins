package awsxray

import (
	"fmt"
	"net/http"
	"time"
)

type segment struct {
	TraceId   string  `json:"trace_id"`
	Id        string  `json:"id"`
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
	Name      string  `json:"name"`
	Type      string  `json:"type,omitempty"`
	ParentId  string  `json:"parent_id,omitempty"`
	HTTP      *HTTP   `json:"http,omitempty"`
	Error     bool    `json:"error,omitempty"`
	Fault     bool    `json:"fault,omitempty"`
}

type HTTP struct {
	Request  *request  `json:"request"`
	Response *response `json:"response"`
}

type request struct {
	Method    string `json:"method,omitempty"`
	URL       string `json:"url,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
	ClientIP  string `json:"client_ip,omitempty"`
}

type response struct {
	Status int `json:"status,omitempty"`
}

func (s *segment) SetStatus(status int) {
	switch {
	case status >= 500:
		s.Fault = true
	case status >= 400:
		s.Error = true
	}
	s.HTTP.Response.Status = status
}

// newHTTP returns a http struct
func newHTTP(r *http.Request) *HTTP {
	scheme := "http"
	host := r.Host

	if len(r.URL.Scheme) > 0 {
		scheme = r.URL.Scheme
	}

	if len(r.URL.Host) > 0 {
		host = r.URL.Host
	}

	return &HTTP{
		Request: &request{
			Method:    r.Method,
			URL:       fmt.Sprintf("%s://%s%s", scheme, host, r.URL.Path),
			ClientIP:  getIp(r),
			UserAgent: r.UserAgent(),
		},
		Response: &response{
			Status: 200,
		},
	}
}

// newSegment creates a new segment based on whether we're part of an existing flow
func newSegment(name string, r *http.Request) *segment {
	// attempt to get IDs first
	parentId := getParentId(r.Header)
	traceId := getTraceId(r.Header)

	// now set the trace ID
	traceHdr := r.Header.Get(TraceHeader)
	traceHdr = setTraceId(traceHdr, traceId)

	// create segment
	s := &segment{
		Id:        getRandom(8),
		HTTP:      newHTTP(r),
		Name:      name,
		TraceId:   traceId,
		StartTime: float64(time.Now().Truncate(time.Millisecond).UnixNano()) / 1e9,
	}

	// if we have a parent then we are a subsegment
	if len(parentId) > 0 {
		s.ParentId = parentId
		s.Type = "subsegment"
	} else {
		// set a new parent Id
		traceHdr = setParentId(traceHdr, s.Id)
	}

	// now save the header for the future context
	r.Header.Set(TraceHeader, traceHdr)

	return s
}
