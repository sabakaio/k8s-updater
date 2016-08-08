package updater

import (
	"fmt"
	"github.com/blang/semver"
	"github.com/sabakaio/k8s-updater/pkg/registry"
	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
	"strings"
)

// GetName return container name
func (c *Container) GetName() string {
	return c.container.Name
}

// ParseImageVersion return semver of current container image
func (c *Container) ParseImageVersion() (semver.Version, error) {
	image := strings.Split(c.container.Image, ":")
	if len(image) != 2 {
		return semver.Version{}, fmt.Errorf(
			"invalid image name, could not extract version: %s", c.container.Image)
	}
	return semver.ParseTolerant(image[1])
}

// GetDockerRegistry returns the container image registry
func (c *Container) GetDockerRegistry() (*registry.Registry, error) {
	return registry.NewRegistry(c.deployment)
}

// GetLatestVersion returns a latest image version from repository
func (c *Container) GetLatestVersion() (semver.Version, error) {
	image := strings.Split(c.container.Image, ":")
	registry, err := c.GetDockerRegistry()
	if err != nil {
		return semver.Version{}, err
	}
	return registry.GetLatestVersion(image[0])
}

// NewList list containers to check for updates
func NewList(k *client.Client, namespace string) (containers *ContainerList, err error) {
	selector, err := labels.Parse("autoupdate")
	if err != nil {
		return
	}

	opts := api.ListOptions{
		LabelSelector: selector,
	}
	deployments, err := k.Deployments(namespace).List(opts)
	if err != nil {
		return
	}
	containers = new(ContainerList)
	for _, d := range deployments.Items {
		for _, c := range d.Spec.Template.Spec.Containers {
			var container = &Container{
				container:  &c,
				deployment: &d,
			}
			containers.Items = append(containers.Items, container)
		}
	}

	return
}
