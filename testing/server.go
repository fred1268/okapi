package testing

// AuthenticationAPIKey represents the API Key
// authentication parameters.
type AuthenticationAPIKey struct {
	// APIKey represents the API Key itself.
	APIKey string
	// Header represents the header used to
	// set the API Key.
	Header string
}

// Authentication represents the authentication mode.
type Authentication struct {
	// Login represents a login/password authentication.
	Login *APIRequest
	// APIKey represents an authentication with API keys.
	APIKey *AuthenticationAPIKey
}

// Session represents how the session is maintained.
type Session struct {
	// Cookie represents the name of the cookie.
	Cookie string
	// JWT represents how the JWT is provided to the client.
	JWT string
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
	// Session represents how the session is maintained
	// by the server.
	Session *Session
	// UserAgent represents the user agent okapi uses.
	// The UserAgent field has meaningful default.
	UserAgent string
	// Timeout represents the timeout used in every request.
	// The Timeout field has a meaningful default.
	Timeout int
}
