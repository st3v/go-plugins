package awsxray

import (
	"fmt"
	"time"

	"github.com/micro/go-micro/metadata"
	"golang.org/x/net/context"
)

type contextSegmentKey struct{}

type segment struct {
	TraceId   string  `json:"trace_id"`
	Id        string  `json:"id"`
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
	Name      string  `json:"name"`
	Type      string  `json:"type,omitempty"`
	ParentId  string  `json:"parent_id,omitempty"`
	HTTP      http    `json:"http,omitempty"`
}

type http struct {
	Request  request  `json:"request"`
	Response response `json:"response"`
}

type request struct {
	Method string `json:"method,omitempty"`
	URL    string `json:"url,omitempty"`
}

type response struct {
	Status int `json:"status,omitempty"`
}

// getSegment creates a new segment based on whether we're part of an existing flow
func getSegment(name string, ctx context.Context) *segment {
	var parentId string

	// try get existing segment for parent Id
	if p, ok := ctx.Value(contextSegmentKey{}).(*segment); ok {
		parentId = p.Id
	}

	// get metadata
	md, _ := metadata.FromContext(ctx)

	// create segment
	s := &segment{
		Id:        fmt.Sprintf("%x", getRandom(8)),
		Name:      name,
		TraceId:   getTraceId(md),
		StartTime: float64(time.Now().Truncate(time.Millisecond).UnixNano()) / 1e9,
	}

	// we have a parent so subsegment
	if len(parentId) > 0 {
		s.ParentId = parentId
		s.Type = "subsegment"
	}

	return s
}
