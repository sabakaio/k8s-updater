package registry

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/blang/semver"
	ext "k8s.io/kubernetes/pkg/apis/extensions"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"net/http"
	"strings"
)

// NewRegistry create a new registry with given name and credentials and configure http client
func NewRegistry(name string, credentials *Credentials) (r *Registry, err error) {
	transport := &BasicAuthTransport{
		Username: credentials.Username,
		Password: credentials.Password,
	}
	c := &http.Client{Transport: transport}
	r = &Registry{Name: name, credentials: credentials, client: c}
	return
}

// GetHost return the registry's host for API requests
func (r *Registry) GetHost() string {
	if r.Name == "default" {
		return "index.docker.io"
	}
	return r.Name
}

// Get makes GET API call for given path with auth headers
func (r *Registry) Get(pathTemplate string, args ...interface{}) (response *http.Response, err error) {
	path := fmt.Sprintf(pathTemplate, args...)
	url := fmt.Sprintf("https://%s%s", r.GetHost(), path)
	response, err = r.client.Get(url)
	return
}

// GetTags returns list of tags for the image repository
func (r *Registry) GetTags(repo string) (tags []string, err error) {
	res, err := r.Get("/v2/%s/tags/list", repo)
	defer res.Body.Close()

	if err != nil {
		return
	}
	if res.StatusCode >= 400 {
		err = fmt.Errorf("cannot get tag list for '%s': %s", repo, res.Status)
		return
	}

	target := new(TagsList)
	err = json.NewDecoder(res.Body).Decode(target)
	if err != nil {
		return
	}
	tags = target.Tags
	return
}

// GetRegistries returns list of registries based on deployment's image pull secrets
func GetRegistries(k *client.Client, deployment *ext.Deployment) (registries *RegistryList, err error) {
	registries = new(RegistryList)
	for _, s := range deployment.Spec.Template.Spec.ImagePullSecrets {
		// Get image pull secret, ...
		secret, e := k.Secrets(deployment.Namespace).Get(s.Name)
		if e != nil {
			err = e
			return
		}
		// ... decode from base64, ...
		cfg_b64 := b64.StdEncoding.EncodeToString(secret.Data[".dockercfg"])
		cfg, e := b64.StdEncoding.DecodeString(cfg_b64)
		if e != nil {
			err = e
			return
		}
		// ... and unmarshal into map of credentials structs
		var credentialsMap map[string]*Credentials
		err = json.Unmarshal(cfg, &credentialsMap)
		if err != nil {
			return
		}

		for name, creds := range credentialsMap {
			if r, e := NewRegistry(name, creds); e == nil {
				registries.Items = append(registries.Items, r)
			} else {
				err = e
				return
			}
		}
	}
	return
}

// Get returns a registry with the given name from the list
func (list *RegistryList) Get(name string) (r *Registry, err error) {
	for _, item := range list.Items {
		if item.Name == name {
			r = item
			return
		}
	}
	err = fmt.Errorf("registry with name '%s' is not in the list", name)
	return
}

// NewRepository return a new Repository for given image name
func NewRepository(image string, registry *Registry) *Repository {
	// Cut off image version
	image = strings.SplitN(image, ":", 2)[0]
	// Cut off custom registry domain from image name
	if strings.HasPrefix(image, registry.Name+"/") {
		image = strings.SplitN(image, "/", 2)[1]
	}
	return &Repository{
		Name:     image,
		Registry: registry,
	}
}

// GetLatestVersion returns the latest image version based on tag
func (r *Repository) GetLatestVersion() (version semver.Version, err error) {
	tags, err := r.Registry.GetTags(r.Name)
	if err != nil {
		return
	}
	if len(tags) == 0 {
		err = fmt.Errorf("there is no image tags for '%s'", r.Name)
	}
	for _, tag := range tags {
		v, e := semver.ParseTolerant(tag)
		if e == nil && v.GT(version) {
			version = v
		}
	}
	return
}
