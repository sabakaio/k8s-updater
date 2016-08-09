package updater

import (
	"github.com/sabakaio/k8s-updater/pkg/registry"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/batch"
	ext "k8s.io/kubernetes/pkg/apis/extensions"
)

// Container holds a container to check for version update linked with `Deployment`
type Container struct {
	container    *api.Container
	deployment   *ext.Deployment
	repository   *registry.Repository
	beforeUpdate *batch.Job
}

// ContainerList is a list of containers to check for version update
type ContainerList struct {
	Items []*Container
}
