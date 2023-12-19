package testing

import (
	"fmt"
	"strings"

	"github.com/fred1268/okapi/testing/internal/os"
)

// APIRequest contains all information required to run a test.
type APIRequest struct {
	// Name represents the name of the test.
	// It should be globally unique.
	Name string
	// Server represents the name of the server
	// as written in the servers configuration file.
	Server string
	// Method represents the HTTP Method.
	// See http.MethodGet, etc.
	Method string
	// Endpoint represents your ReST API endpoint.
	// It should be relative to the server's Host.
	Endpoint string
	// Headers represents additional headers to add
	// to the request.
	Headers map[string]string
	// URLParams represents additional query parameters
	// that will be send with the request.
	URLParams map[string]string
	// Payload represents the payload provided with some
	// methods (POST, PUT, etc.) to the request.
	Payload string
	// Expected represents the expected APIResponse if
	// everything goes according to the plan. The Logs
	// field is ignored in this context.
	Expected *APIResponse
	// Capture allows okapi to capture the response as
	// a JSON object and make it available for the next
	// tests (fileParallel modes only).
	Capture bool
	// CaptureJWT allows okapi to update the current JWT
	// and make it available for the next requests.
	CaptureJWT bool
	// Skip will make okapi skip this test. Can use useful
	// when debugging script files or to allow tests to
	// pass while a bug is being fixed for instance.
	Skip bool
	// Debug will make okapi output test debugging
	// information to ease troubleshooting errors
	Debug  bool
	atFile bool
}

// APIResponse contains information about the response from
// the server.
type APIResponse struct {
	// StatusCode represents the HTTP Status Code returned
	// by the server.
	StatusCode int
	// Response represents the payload (response) returned
	// by the server.
	Response string
	// Logs represents okapi's logs which are grouped later
	// on to be nicely displayed even in parallel mode.
	Logs   []string
	atFile bool
}

func (a *APIRequest) validate() error {
	if strings.Contains(a.Name, ".") {
		return fmt.Errorf("name cannot contain the . (period) character")
	}
	if a.Server == "" && !strings.Contains(a.Endpoint, "://") {
		return fmt.Errorf("empty server or relative endpoint")
	}
	if a.Method == "" || a.Endpoint == "" || a.Expected == nil {
		return fmt.Errorf("empty method, endpoint or expectations")
	}
	a.Endpoint = os.SubstituteEnvironmentVariable(a.Endpoint)
	a.Payload = os.SubstituteEnvironmentVariable(a.Payload)
	return nil
}

func (a *APIRequest) hasFileDepencies() bool {
	return a.atFile || a.Expected.atFile
}
