package updater

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/blang/semver"
	"github.com/sabakaio/k8s-updater/pkg/registry"
	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
	"strings"
)

// GetName returns the container name
func (c *Container) GetName() string {
	return c.container.Name
}

// GetImageName returns the container image name
func (c *Container) GetImageName() string {
	image := strings.Split(c.container.Image, ":")
	return image[0]
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

// GetLatestVersion returns a latest image version from repository
func (c *Container) GetLatestVersion() (semver.Version, error) {
	image := strings.Split(c.container.Image, ":")
	return c.registry.GetLatestVersion(image[0])
}

// NewList list containers to check for updates
func NewList(k *client.Client, namespace string) (containers *ContainerList, err error) {
	// List all deployments lebled with `autoupdate`
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

	// Iterate over the list of deployments to get a list of containers
	containers = new(ContainerList)
	for _, d := range deployments.Items {
		// Get deployment spec registries
		registries, e := registry.GetRegistries(k, &d)
		if e != nil {
			err = e
			return
		}
		// Iterate over pod containers to get update targets
		for _, c := range d.Spec.Template.Spec.Containers {
			var container = &Container{
				container:  &c,
				deployment: &d,
			}
			// Choose a registry for container by the name
			for _, r := range registries.Items {
				if strings.HasPrefix(container.GetImageName(), r.Name+"/") {
					container.registry = r
					break
				}
			}
			if container.registry == nil {
				if defaultRegistry, err := registries.Get("default"); err != nil {
					container.registry = defaultRegistry
				} else {
					log.Error("container '%s' of deployment '%s' has no private registry configured", c.Name, d.Name)
					continue
				}
			}
			log.Debugf("container '%s' of deployment '%s' uses '%s' registry", c.Name, d.Name, container.registry.Name)
			containers.Items = append(containers.Items, container)
		}
	}

	return
}
