package watch

import (
	"bufio"
	"encoding/json"
	"net/http"
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

	go func() {
		for scanner.Scan() {
			var event Event
			err := json.Unmarshal(scanner.Bytes(), &event)
			if err == nil {
				wr.results <- event
			}
		}
		wr.Stop()
	}()
}

// NewBodyWatcher creates a k8s body watcher for
// a given http request
func NewBodyWatcher(req *http.Request) (Watch, error) {
	stop := make(chan struct{})
	req.Cancel = stop

	res, err := http.DefaultClient.Do(req)
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
