package updater

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/blang/semver"
	"github.com/sabakaio/k8s-updater/pkg/registry"
	"github.com/sabakaio/k8s-updater/pkg/util"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/batch"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
	"strings"
	"time"
)

// GetDeploymentName returns the deployment name
func (c *Container) GetDeploymentName() string {
	return c.deployment.Name
}

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

// SetImageVersion updates Deployment template with the set version. It does not save the deployment.
// NOTE a version is passed by the value to avoid nil pointer errors
func (c *Container) SetImageVersion(v registry.Version) (*Container, error) {
	image := strings.SplitN(c.GetImageName(), ":", 2)
	newImage := image[0] + ":" + v.String()
	c.container.Image = newImage
	for i, dc := range c.deployment.Spec.Template.Spec.Containers {
		if dc.Name == c.GetName() {
			c.deployment.Spec.Template.Spec.Containers[i].Image = newImage
		}
	}
	return c, nil
}

// GetLatestVersion returns a latest image version from repository
func (c *Container) GetLatestVersion() (*registry.Version, error) {
	return c.repository.GetLatestVersion()
}

// GetAutoupdateVersion returns version to perform autoupdate to.
// nil will be returned if notheng to update.
func (c *Container) GetAutoupdateVersion() (version *registry.Version, err error) {
	current, err := c.GetImageVersion()
	if err != nil {
		return
	}
	latest, err := c.GetLatestVersion()
	if err != nil {
		return
	}
	// Update container deployment if greater image version found
	if latest.Semver.GT(current.Semver) {
		version = latest
	}
	return
}

// SetRepositoryFrom iterate over registries list to match containers image repository
func (c *Container) SetRepositoryFrom(registries *registry.RegistryList) error {
	image := c.GetImageName()
	// Choose a registry for container by the name
	for _, r := range registries.Items {
		if strings.HasPrefix(image, r.Name+"/") {
			c.repository = registry.NewRepository(image, r)
			return nil
		}
	}
	if defaultRegistry, err := registries.Get("default"); err == nil {
		c.repository = registry.NewRepository(image, defaultRegistry)
		return nil
	}
	return fmt.Errorf("cannot match registry for container '%s'", c.GetName())
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

		// Get deployment annotations to use as config for updater
		annotations := d.GetAnnotations()

		// Iterate over pod containers to get update targets
		for _, c := range d.Spec.Template.Spec.Containers {
			var container = &Container{
				container:  &c,
				deployment: &d,
			}

			if err := container.SetRepositoryFrom(registries); err != nil {
				log.Errorln(err.Error())
				continue
			}
			log.Debugf("container '%s' of deployment '%s' uses '%s' repository", c.Name, d.Name, container.repository.Name)

			hook_key := "before_autoupdate_" + c.Name
			hook_name, ok := annotations[hook_key]
			if ok {
				job, e := k.Batch().Jobs(d.Namespace).Get(hook_name)
				if e == nil {
					container.beforeUpdate = job
					log.Debugf("before update hook for '%s' container: %s", c.Name, job.Name)
				}
			}

			containers.Items = append(containers.Items, container)
		}
	}

	return
}

// UpdateDeployment updates Deployment version on the cluster
func (c *Container) UpdateDeployment(k *client.Client, v registry.Version) (err error) {
	namespace := c.deployment.Namespace

	// Update image version for container
	newContainer, err := c.SetImageVersion(v)
	if err != nil {
		return err
	}

	// Perform preupdate hook if exists
	hook := newContainer.GetBeforeUpdateJob()
	if hook != nil {
		// Create a job by hook spec
		createdJob, e := k.Batch().Jobs(namespace).Create(hook)
		if e != nil {
			err = e
			return
		}
		job_name := createdJob.GetName()
		log.Debugln("before update hook job created with name", job_name)

		// Defer job delete
		defer func() {
			if e := util.DeletePodsInJob(k, createdJob); e != nil {
				log.Errorf("could not delete pods related to job: %s", e.Error())
			}
			deleteOptions := &api.DeleteOptions{}
			if e := k.Batch().Jobs(namespace).Delete(job_name, deleteOptions); e != nil {
				log.Errorf("could not cleanup job: %s", e.Error())
			} else {
				log.Debugf("job %s deleted", job_name)
			}
		}()

		// Wait the job to complete
		for {
			runningJob, e := k.Batch().Jobs(namespace).Get(job_name)
			if e != nil {
				err = e
				return
			}
			if runningJob.Status.Failed > 0 {
				err = fmt.Errorf("job %s failed")
				return
			}
			if runningJob.Status.Succeeded > 0 {
				break
			}
			time.Sleep(time.Second * 3)
		}
	}

	_, err = k.Deployments(namespace).Update(newContainer.deployment)
	// TODO: what to do with the new deployment? Update our memory store?

	return
}

// GetBeforeUpdateJob returns configuration for a job to run before deployment update.
// It copies the original job and fix it's container image version.
func (c *Container) GetBeforeUpdateJob() (job *batch.Job) {
	if c.beforeUpdate == nil {
		return
	}
	hook := c.beforeUpdate

	job = &batch.Job{}
	job.Spec.Template.Spec = hook.Spec.Template.Spec
	for i, jobContainer := range job.Spec.Template.Spec.Containers {
		jobImage := strings.SplitN(jobContainer.Image, ":", 2)[0]
		if strings.HasPrefix(c.GetImageName(), jobImage+":") {
			job.Spec.Template.Spec.Containers[i].Image = c.GetImageName()
			log.Debugln("update job container image to", c.GetImageName())
		}
	}

	genName := c.GetName() + "-" + hook.GetName() + "-"
	job.ObjectMeta.SetGenerateName(genName)

	return
}
