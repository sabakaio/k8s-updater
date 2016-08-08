package registry

import (
	"github.com/blang/semver"
	ext "k8s.io/kubernetes/pkg/apis/extensions"
)

func NewRegistry(deployment *ext.Deployment) (r *Registry, err error) {
	return
}

func (r *Registry) GetLatestVersion(repo string) (v semver.Version, err error) {
	return semver.Make("100.0.0")
}
