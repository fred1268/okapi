package testing

import (
	"fmt"
	"strings"
)

// AuthenticationAPIKey represents the API Key
// authentication parameters.
type AuthenticationAPIKey struct {
	// APIKey represents the API Key itself.
	APIKey string
	// Header represents the header used to
	// set the API Key.
	Header string
}

// Session represents how the session is maintained.
type Session struct {
	// Cookie represents the name of the cookie.
	Cookie string
	// JWT represents how the JWT is provided to the client.
	JWT string
}

// Authentication represents the authentication mode.
type Authentication struct {
	// Login represents a login/password authentication.
	Login *APIRequest
	// Session represents how the session is maintained
	// by the server.
	Session *Session
	// APIKey represents an authentication with API keys.
	APIKey *AuthenticationAPIKey
}

// ServerConfig represents a server configuration.
type ServerConfig struct {
	// Host represents the common part of all requests.
	// Usually something like: https://server/
	Host string
	// Headers represents headers to set on each request.
	// The Headers field has meaningful default.
	Headers map[string]string
	// Auth represents the authentication mode.
	Auth *Authentication
	// UserAgent represents the user agent okapi uses.
	// The UserAgent field has meaningful default.
	UserAgent string
	// Timeout represents the timeout used in every request.
	// The Timeout field has a meaningful default.
	Timeout int
}

func (s *ServerConfig) validate() error {
	if s.Auth != nil {
		if s.Auth.Login != nil {
			if err := s.Auth.Login.validate(); err != nil {
				return fmt.Errorf("invalid login information: %w", err)
			}
			if s.Auth.Session == nil || s.Auth.Session.Cookie == "" && s.Auth.Session.JWT == "" {
				return fmt.Errorf("no or invalid session information")
			}
			if s.Auth.Session.JWT != "" && s.Auth.Session.JWT != "header" && !strings.HasPrefix(s.Auth.Session.JWT, "payload") {
				return fmt.Errorf("no or invalid JWT information")
			}
		} else if s.Auth.APIKey == nil {
			return fmt.Errorf("no authentication provided")
		}
	}
	return nil
}
