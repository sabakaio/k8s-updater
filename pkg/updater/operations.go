package updater

import (
	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
)

// GetName return container name
func (c *Container) GetName() string {
	return c.container.Name
}

// ParseImageVersion return semver of current container image
func (c *Container) ParseImageVersion() (version string, err error) {
	version = c.container.Image
	return
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
