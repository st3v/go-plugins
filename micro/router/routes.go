package router

import (
	"math/rand"
	"net/http"
	"time"
)

// Routes is the config expected to be loaded
type Routes struct {
	Routes []Route `json:"routes"`
	// TODO: default route
}

// Route describes a single route which is matched
// on Request and if so, will return the Response
type Route struct {
	Request  Request  `json:"request"`
	Response Response `json:"response"`
	Priority int      `json:"priority"` // 0 is highest. Used for ordering routes
	// TODO: Weight
	// TODO: Type: Proxy, Default
}

// Request describes the expected request and will
// attempt to match all fields specified
type Request struct {
	Method string            `json:"method"`
	Header map[string]string `json:"header"`
	Host   string            `json:"host"`
	Path   string            `json:"path"`
	Query  map[string]string `json:"query"`
	// TODO: RemoteAddr, Body
}

// Response is put into the http.Response for a Request
type Response struct {
	Status     string            `json:"status"`
	StatusCode int               `json:"status_code"`
	Header     map[string]string `json:"header"`
	Body       []byte            `json:"body"`
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (r Route) Match(req *http.Request) bool {
	// bail on nil
	if len(r.Request.Method) == 0 || len(r.Request.Path) == 0 {
		return false
	}

	// just for ease
	rq := r.Request

	// first level match, quick and dirty
	if (rq.Method == req.Method) && (rq.Host == req.Host) && (rq.Path == req.URL.Path) {
		// skip
	} else {
		return false
	}

	// match headers
	for k, v := range rq.Header {
		// does it match?
		if rv := req.Header.Get(k); rv != v {
			return false
		}
	}

	// match query
	vals := req.URL.Query()
	for k, v := range rq.Query {
		// does it match?
		if rv := vals.Get(k); rv != v {
			return false
		}
	}

	// we got a match!
	return true
}

func (r Route) Write(w http.ResponseWriter, req *http.Request) {
	// set headers
	for k, v := range r.Response.Header {
		w.Header().Set(k, v)
	}
	// set status code
	w.WriteHeader(r.Response.StatusCode)

	// set response
	if len(r.Response.Body) > 0 {
		w.Write(r.Response.Body)
	} else {
		w.Write([]byte(r.Response.Status))
	}
}
