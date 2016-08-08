package registry

// Registry is a docker registry with `Name` and private `credentials`
type Registry struct {
	Name        string
	credentials *Credentials
}

// Credentials is a structure to unmarshall .dockercfg data
type Credentials struct {
	username string
	password string
	email    string
	auth     string
}

type RegistryList struct {
	Items []*Registry
}
