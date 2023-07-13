package testing

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/fred1268/okapi/testing/internal/json"
)

type Client struct {
	config *ServerConfig
	client *http.Client
	cookie *http.Cookie
	jwt    string
}

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

func (c *Client) getRequest(ctx context.Context, apiRequest *APIRequest) (*http.Request, error) {
	addr, err := c.buildEndpointURL(ctx, apiRequest)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(apiRequest.Method), addr,
		bytes.NewBufferString(apiRequest.Payload))
	if err != nil {
		return nil, err
	}
	if c.config.Auth != nil && c.config.Auth.APIKey != nil {
		req.Header.Set(c.config.Auth.APIKey.Header, c.config.Auth.APIKey.APIKey)
	}
	req.Header.Add("User-Agent", c.config.UserAgent)
	if c.jwt != "" {
		req.Header.Set("authorization", fmt.Sprintf("Bearer: %s", c.jwt))
	}
	if c.cookie != nil {
		req.AddCookie(c.cookie)
	}
	if len(apiRequest.Headers) != 0 {
		for key, value := range apiRequest.Headers {
			req.Header.Add(key, value)
		}
	} else {
		for key, value := range c.config.Headers {
			req.Header.Add(key, value)
		}
	}
	return req, nil
}

func (c *Client) call(ctx context.Context, apiRequest *APIRequest) (response *APIResponse, err error) {
	var req *http.Request
	req, err = c.getRequest(ctx, apiRequest)
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
	if c.cookie == nil && c.config.Session != nil && c.config.Session.Cookie.Name != "" {
		cookies := resp.Cookies()
		for _, cookie := range cookies {
			if cookie.Name == c.config.Session.Cookie.Name {
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
	if c.jwt == "" && c.config.Session != nil {
		switch c.config.Session.JWT {
		case "payload":
			// TODO use something like {token:"..."}
			c.jwt = string(res)
		case "header":
			c.jwt = resp.Header.Get("authorization")
		default:
			err = ErrInvalidJWTSession
		}
	}
	response = &APIResponse{StatusCode: resp.StatusCode, Response: string(res)}
	return
}

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

func (c *Client) Test(ctx context.Context, apiRequest *APIRequest, verbose bool) (response *APIResponse, err error) {
	start := time.Now()
	defer func() {
		if err == nil {
			if verbose {
				response.Logs = append(response.Logs, fmt.Sprintf("    --- PASS:\t%s (%0.2fs)\n", apiRequest.Name,
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
	response, err = c.call(ctx, apiRequest)
	if err != nil {
		return
	}
	if response.StatusCode != apiRequest.Expected.StatusCode {
		err = ErrStatusCodeMismatched
		return
	}
	err = json.CompareJSONStrings(ctx, apiRequest.Expected.Response, response.Response)
	if errors.Is(err, json.ErrJSONMismatched) {
		err = errors.Join(err, ErrResultMismatched)
	}
	return
}
