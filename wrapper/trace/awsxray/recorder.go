package awsxray

import (
	"encoding/json"
	"net"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/xray"
)

type recorder struct {
	opts Options
}

func (r recorder) record(s *segment) error {
	// set end time
	s.EndTime = float64(time.Now().Truncate(time.Millisecond).UnixNano()) / 1e9

	// marshal
	b, _ := json.Marshal(s)

	// Use XRay Client if available
	if r.opts.Client != nil {
		_, err := r.opts.Client.PutTraceSegments(&xray.PutTraceSegmentsInput{
			TraceSegmentDocuments: []*string{
				aws.String("TraceSegmentDocument"),
				aws.String(string(b)),
			},
		})
		return err
	}

	// Use Daemon
	c, err := net.Dial("udp", r.opts.Daemon)
	if err != nil {
		return err
	}

	header := append([]byte(`{"format": "json", "version": 1}`), byte('\n'))
	_, err = c.Write(append(header, b...))
	return err
}
