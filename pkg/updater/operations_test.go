package updater

import (
	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/kubernetes/pkg/api"
	"testing"
)

func TestContainer(t *testing.T) {
	k8sContainer := api.Container{
		Name:  "my-container",
		Image: "registry.example.com/my-image:1.2.3",
	}
	container := Container{
		container: &k8sContainer,
	}
	Convey("Test Container type", t, func() {
		So(container.GetName(), ShouldEqual, "my-container")
	})
}
