package updater

import (
	"fmt"
	"github.com/blang/semver"
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
