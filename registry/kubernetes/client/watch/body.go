package watch

import (
	"bufio"
	"encoding/json"
	"net/http"
	"time"
)

// bodyWatcher scans the body of a request for chunks
type bodyWatcher struct {
	results chan Event
	stop    chan struct{}
	res     *http.Response
	req     *http.Request
}

// Changes returns the results channel
func (wr *bodyWatcher) ResultChan() <-chan Event {
	return wr.results
}

// Stop cancels the request
func (wr *bodyWatcher) Stop() {
	select {
	case <-wr.stop:
		return
	default:
		close(wr.stop)
		close(wr.results)
	}
}

func (wr *bodyWatcher) stream() {
	scanner := bufio.NewScanner(wr.res.Body)

	// ignore first few messages from stream,
	// as they are usually old.
	ignore := true
	go func() {
		<-time.After(time.Second)
		ignore = false
	}()

	go func() {
		for scanner.Scan() {
			if ignore {
				continue
			}

			var event Event
			err := json.Unmarshal(scanner.Bytes(), &event)
			if err != nil {
				continue
			}
			wr.results <- event
		}
		wr.Stop()
	}()
}

// NewBodyWatcher creates a k8s body watcher for
// a given http request
func NewBodyWatcher(req *http.Request, client *http.Client) (Watch, error) {
	stop := make(chan struct{})
	req.Cancel = stop

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	wr := &bodyWatcher{
		results: make(chan Event),
		stop:    stop,
		req:     req,
		res:     res,
	}

	go wr.stream()
	return wr, nil
}
