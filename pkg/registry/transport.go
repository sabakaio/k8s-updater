package registry

import (
	"net/http"
)

type BasicAuthTransport struct {
	Username string
	Password string
}

func (t *BasicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Username != "" || t.Password != "" {
		req.SetBasicAuth(t.Username, t.Password)
	}
	return http.DefaultTransport.RoundTrip(req)
}
