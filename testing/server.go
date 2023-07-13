package testing

type AuthenticationAPIKey struct {
	APIKey string
	Header string
}

type Authentication struct {
	Login  *APIRequest
	APIKey *AuthenticationAPIKey
}

type SessionCookie struct {
	Name string
}

type Session struct {
	Cookie *SessionCookie
	JWT    string
}

type ServerConfig struct {
	Host      string
	Headers   map[string]string
	Auth      *Authentication
	Session   *Session
	UserAgent string
	Timeout   int
}
