package atto

import "net/http"

// The DefaultNodeAuthenticator is the NodeAuthenticator that is used to
// authenticate all HTTP requests that are sent to a node. If it is nil,
// no authentication will be performed.
var DefaultNodeAuthenticator NodeAuthenticator

// A NodeAuthenticator can add authentication credentials to an HTTP
// request.
type NodeAuthenticator interface {
	Authenticate(request *http.Request)
}

// NodeBasicAuth holds HTTP Basic Authentication credentials.
type NodeBasicAuth struct {
	Username string // The HTTP Basic Authentication username.
	Password string // The HTTP Basic Authentication password.
}

// Authenticate adds the HTTP Basic Authentication headers to request.
func (auth NodeBasicAuth) Authenticate(request *http.Request) {
	request.SetBasicAuth(auth.Username, auth.Password)
}
