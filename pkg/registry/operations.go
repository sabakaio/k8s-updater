package registry

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/blang/semver"
	ext "k8s.io/kubernetes/pkg/apis/extensions"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

// GetGetRegistries returns list of registries based on deployment's image pull secrets
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
			r := &Registry{Name: name, credentials: creds}
			registries.Items = append(registries.Items, r)
		}
	}
	return
}

func (r *Registry) GetLatestVersion(repo string) (v semver.Version, err error) {
	return semver.Make("100.0.0")
}

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
