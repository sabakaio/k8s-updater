package registry

import (
	"net/http"
)

// Registry is a docker registry with `Name` and private `credentials`
type Registry struct {
	Name        string
	credentials *Credentials
	client      *http.Client
}

// Credentials is a structure to unmarshall .dockercfg data
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Auth     string `json:"auth"`
}

type RegistryList struct {
	Items []*Registry
}

// Repository is an image repository in a registry
type Repository struct {
	Name     string
	Registry *Registry
}

// TagList is a get tags API call response
type TagsList struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}
