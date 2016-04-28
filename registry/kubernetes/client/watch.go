package client

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// WatchRequest listens for changes on a given k8s resource
type WatchRequest struct {
	change chan *Event
	stop   chan bool
}

// Change returns the change channel
func (wr *WatchRequest) Change() chan *Event {
	return wr.change
}

// Stop cancels the request
func (wr *WatchRequest) Stop() {
	close(wr.stop)
}

func (wr *WatchRequest) stream(body io.ReadCloser) {
	scanner := bufio.NewScanner(body)
	defer body.Close()

	for scanner.Scan() {
		select {
		default:
			var data Event
			err := json.Unmarshal(scanner.Bytes(), &data)

			if err == nil {
				wr.change <- &data
			} else {
				log.Printf("unmarshal error on watch %v", err)
			}

		case <-wr.stop:
			return
		}
	}
}

// newWatchRequest creates a new WatchRequest
func newWatchRequest(k *Client, url string) (*WatchRequest, error) {
	wr := &WatchRequest{
		change: make(chan *Event),
		stop:   make(chan bool),
	}

	url += "?watch=true"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := k.client.Do(req)
	if err != nil {
		return nil, err
	}

	go wr.stream(res.Body)

	return wr, nil
}
