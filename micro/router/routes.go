package router

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
	// TODO: weight
}

// Request describes the expected request and will
// attempt to match all fields specified
type Request struct {
	Method string            `json:"string"`
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
