package testing

type APIRequest struct {
	Name      string
	Server    string
	Method    string
	Endpoint  string
	Headers   map[string]string
	URLParams map[string]string
	Payload   string
	Expected  *APIResponse
}

type APIResponse struct {
	StatusCode int
	Response   string
	Logs       []string
}
