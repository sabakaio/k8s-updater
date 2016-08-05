package updater

import (
	"k8s.io/kubernetes/pkg/api"
	ext "k8s.io/kubernetes/pkg/apis/extensions"
)

// Container hold a container to check for version update linked with `Deployment`
type Container struct {
	container  *api.Container
	deployment *ext.Deployment
}

// List of `Container`s to check for version update
type ContainerList struct {
	Items []*Container
}
