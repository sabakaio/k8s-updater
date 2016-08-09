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
	return c.container.Image
}

// GetImageVersion returns `registry.Version` for current container image
func (c *Container) GetImageVersion() (version *registry.Version, err error) {
	image := strings.SplitN(c.GetImageName(), ":", 2)
	if len(image) != 2 {
		err = fmt.Errorf("invalid image name, could not extract version: %s", c.GetImageName())
		return
	}
	semver, err := semver.ParseTolerant(image[1])
	if err != nil {
		return
	}
	version = &registry.Version{
		Tag:    image[1],
		Semver: semver,
	}
	return
}

// UpdateImageVersion updates Deployment template with the set version. It does nto save the deployment.
func (c *Container) UpdateImageVersion(v semver.Version) (*Container, error) {
	image := strings.Split(c.container.Image, ":")
	if len(image) != 2 {
		return c, fmt.Errorf(
			"invalid image name, could not extract version: %s", c.container.Image)
	}
	stringVersion := v.String()
	imageString := strings.Join([]string{image[0], stringVersion}, ":")
	c.container.Image = imageString
	return c, nil
}

// GetLatestVersion returns a latest image version from repository
func (c *Container) GetLatestVersion() (*registry.Version, error) {
	return c.repository.GetLatestVersion()
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
			image := container.GetImageName()
			// Choose a registry for container by the name
			for _, r := range registries.Items {
				if strings.HasPrefix(image, r.Name+"/") {
					container.repository = registry.NewRepository(image, r)
					break
				}
			}
			if container.repository == nil {
				if defaultRegistry, err := registries.Get("default"); err == nil {
					container.repository = registry.NewRepository(image, defaultRegistry)
				} else {
					log.Errorf("container '%s' of deployment '%s' has no private registry configured", c.Name, d.Name)
					continue
				}
			}
			log.Debugf("container '%s' of deployment '%s' uses '%s' repository", c.Name, d.Name, container.repository.Name)
			containers.Items = append(containers.Items, container)
		}
	}

	return
}

// UpdateDeployment updates Deployment version on the cluster
func (c *Container) UpdateDeployment(k *client.Client, namespace string, v semver.Version) error {
	newContainer, err := c.UpdateImageVersion(v)
	if err != nil {
		return err
	}

	newDeployment, err := k.Deployments(namespace).Update(newContainer.deployment)
	fmt.Println(newDeployment.GetName())
	// TODO: what to do with the new deployment? Update our memory store?

	if err != nil {
		return err
	}

	return nil
}
