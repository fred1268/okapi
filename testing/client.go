package testing

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	ijson "github.com/fred1268/okapi/testing/internal/json"
)

// Client represent an API Client for the specified server.
type Client struct {
	config *ServerConfig
	client *http.Client
	cookie *http.Cookie
	jwt    string
}

// NewClient returns a new client according to the
// provided ServerConfig.
//
// Client is safe to be used concurrently since
// http.Client is. However, Connect() should only
// be called once.
func NewClient(config *ServerConfig) *Client {
	client := &Client{
		config: config,
		client: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 10 * time.Second,
				MaxIdleConns:          100,
				MaxConnsPerHost:       100,
				MaxIdleConnsPerHost:   100,
			},
		},
	}
	return client
}

func (c *Client) Clone() *Client {
	cookie := http.Cookie{}
	if c.cookie != nil {
		cookie = *c.cookie
	}
	return &Client{
		config: c.config,
		cookie: &cookie,
		jwt:    c.jwt,
		client: &http.Client{
			Timeout: time.Duration(c.config.Timeout) * time.Second,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 10 * time.Second,
				MaxIdleConns:          100,
				MaxConnsPerHost:       100,
				MaxIdleConnsPerHost:   100,
			},
		},
	}
}

func (c *Client) buildEndpointURL(ctx context.Context, apiRequest *APIRequest) (string, error) {
	var err error
	addr := apiRequest.Endpoint
	if !strings.Contains(apiRequest.Endpoint, "://") {
		addr, err = url.JoinPath(c.config.Host, apiRequest.Endpoint)
		if err != nil {
			return "", err
		}
	}
	if len(apiRequest.URLParams) != 0 {
		first := true
		for key, value := range apiRequest.URLParams {
			if first {
				addr = fmt.Sprintf("%s?%s=%s", addr, url.QueryEscape(key), url.QueryEscape(value))
				first = false
			} else {
				addr = fmt.Sprintf("%s&%s=%s", addr, url.QueryEscape(key), url.QueryEscape(value))
			}
		}
	}
	return addr, nil
}

func (c *Client) getRequest(ctx context.Context, apiRequest *APIRequest, apiResponse *APIResponse) (*http.Request, error) {
	addr, err := c.buildEndpointURL(ctx, apiRequest)
	if err != nil {
		return nil, err
	}
	if apiRequest.Debug {
		apiResponse.Logs = append(apiResponse.Logs, fmt.Sprintf("--- %s\n", apiRequest.Name))
		apiResponse.Logs = append(apiResponse.Logs, "API Request:\n")
		apiResponse.Logs = append(apiResponse.Logs, fmt.Sprintf("  URL: %s\n", addr))
		apiResponse.Logs = append(apiResponse.Logs, fmt.Sprintf("  Method: %s\n", apiRequest.Method))
		apiResponse.Logs = append(apiResponse.Logs, fmt.Sprintf("  Payload: %s\n", apiRequest.Payload))
		apiResponse.Logs = append(apiResponse.Logs, "  Headers:\n")
	}
	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(apiRequest.Method), addr,
		bytes.NewBufferString(apiRequest.Payload))
	if err != nil {
		return nil, err
	}
	if c.config.Auth != nil && c.config.Auth.APIKey != nil {
		if apiRequest.Debug {
			apiResponse.Logs = append(apiResponse.Logs, fmt.Sprintf("    %s: %s\n", c.config.Auth.APIKey.Header,
				c.config.Auth.APIKey.APIKey))
		}
		req.Header.Set(c.config.Auth.APIKey.Header, c.config.Auth.APIKey.APIKey)
	}
	if c.config.UserAgent != "" {
		if apiRequest.Debug {
			apiResponse.Logs = append(apiResponse.Logs, fmt.Sprintf("    User-Agent: %s\n", c.config.UserAgent))
		}
		req.Header.Add("User-Agent", c.config.UserAgent)
	}
	if c.jwt != "" {
		if apiRequest.Debug {
			apiResponse.Logs = append(apiResponse.Logs, fmt.Sprintf("    Authorization: Bearer %s\n", c.jwt))
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.jwt))
	}
	if c.cookie != nil {
		req.AddCookie(c.cookie)
	}
	req.Header.Add("X-okapi-testname", apiRequest.Name)
	if len(apiRequest.Headers) != 0 {
		for key, value := range apiRequest.Headers {
			if apiRequest.Debug {
				apiResponse.Logs = append(apiResponse.Logs, fmt.Sprintf("    %s: %s\n", key, value))
			}
			req.Header.Add(key, value)
		}
	} else {
		for key, value := range c.config.Headers {
			if apiRequest.Debug {
				apiResponse.Logs = append(apiResponse.Logs, fmt.Sprintf("    %s: %s\n", key, value))
			}
			req.Header.Add(key, value)
		}
	}
	return req, nil
}

func (c *Client) captureJWT(response string) error {
	var payload interface{}
	err := json.Unmarshal([]byte(response), &payload)
	if err != nil {
		return err
	}
	if obj, ok := payload.(map[string]any); ok {
		for key, value := range obj {
			if key != strings.ToLower(key) {
				obj[strings.ToLower(key)] = value
				delete(obj, key)
			}
		}
		if c.jwt, ok = obj[c.config.Auth.Session.JWT[8:]].(string); !ok {
			return fmt.Errorf("cannot read JWT from payload")
		}
	}
	return nil
}

func (c *Client) call(ctx context.Context, apiRequest *APIRequest) (apiResponse *APIResponse, err error) {
	apiResponse = &APIResponse{}
	var req *http.Request
	req, err = c.getRequest(ctx, apiRequest, apiResponse)
	if err != nil {
		return
	}
	var resp *http.Response
	resp, err = c.client.Do(req)
	if err != nil {
		return
	}
	defer func() {
		err = resp.Body.Close()
	}()
	if c.cookie == nil && c.config.Auth != nil && c.config.Auth.Session != nil && c.config.Auth.Session.Cookie != "" {
		cookies := resp.Cookies()
		for _, cookie := range cookies {
			if cookie.Name == c.config.Auth.Session.Cookie {
				c.cookie = cookie
				break
			}
		}
	}
	var res []byte
	res, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if apiRequest.Debug {
		apiResponse.Logs = append(apiResponse.Logs, "API Response:\n")
		apiResponse.Logs = append(apiResponse.Logs, fmt.Sprintf("  Response: %s", string(res)))
	}
	if c.jwt == "" && c.config.Auth != nil && c.config.Auth.Session != nil && c.config.Auth.Session.JWT != "" {
		switch c.config.Auth.Session.JWT {
		case "payload":
			c.jwt = string(res)
		case "header":
			c.jwt = resp.Header.Get("authorization")
		default: // "payload.xxx"
			err = c.captureJWT(string(res))
		}
	}
	apiResponse.StatusCode = resp.StatusCode
	apiResponse.Response = string(res)
	return
}

// Connect connects the client to the specified server.
//
// This method should not be called in normal circumpstances:
// prefer LoadClients() instead which will load and connect all
// the clients sequencially. Use Connect() if you want to create
// a client independently from the server configuration file.
func (c *Client) Connect(ctx context.Context) (*APIResponse, error) {
	result, err := c.call(ctx, c.config.Auth.Login)
	if err != nil {
		return nil, err
	}
	if result.StatusCode != c.config.Auth.Login.Expected.StatusCode {
		return result, ErrStatusCodeMismatched
	}
	return result, nil
}

// Test runs the provided test.
//
// If verbose is set to true, the APIResponse will contain
// a more detailed information upon exit. Test can return
// ErrStatusCodeMismatched or ErrResultMismatched if the
// function was run successfully, but the test did not pass.
// It will retuen another error upon unexpected condition.
func (c *Client) Test(ctx context.Context, apiRequest *APIRequest, verbose bool) (response *APIResponse, err error) {
	start := time.Now()
	defer func() {
		if response == nil {
			response = &APIResponse{}
		}
		if err == nil {
			if verbose {
				result := "PASS"
				if apiRequest.Skip {
					result = "SKIP"
				}
				response.Logs = append(response.Logs, fmt.Sprintf("    --- %s:\t%s (%0.2fs)\n", result, apiRequest.Name,
					time.Since(start).Seconds()))
			}
		} else {
			response.Logs = append(response.Logs, fmt.Sprintf("    --- FAIL:\t%s (%0.2fs)\n", apiRequest.Name,
				time.Since(start).Seconds()))
			response.Logs = append(response.Logs, fmt.Sprintf("    wanted: '%s' (%d), got '%s' (%d)\n",
				apiRequest.Expected.Response, apiRequest.Expected.StatusCode, strings.Trim(response.Response, "\n"),
				response.StatusCode))
		}
	}()
	if apiRequest.Skip {
		return
	}
	response, err = c.call(ctx, apiRequest)
	if err != nil {
		return
	}
	if response.StatusCode != apiRequest.Expected.StatusCode {
		err = ErrStatusCodeMismatched
		return
	}
	err = ijson.CompareJSONStrings(apiRequest.Expected.Response, response.Response)
	if errors.Is(err, ijson.ErrJSONMismatched) {
		err = errors.Join(err, ErrResponseMismatched)
	}
	return
}
