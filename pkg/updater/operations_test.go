package updater

import (
	"github.com/blang/semver"
	// "github.com/sabakaio/k8s-updater/pkg/util"
	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/kubernetes/pkg/api"
	ext "k8s.io/kubernetes/pkg/apis/extensions"
	"testing"
)

func TestContainer(t *testing.T) {
	k8sContainer := api.Container{
		Name:  "my-container",
		Image: "registry.example.com/my-image:1.2.3",
	}
	k8sDeployment := ext.Deployment{}
	k8sDeployment.Spec.Template.Spec.Containers = append(k8sDeployment.Spec.Template.Spec.Containers, k8sContainer)
	container := Container{
		container:  &k8sContainer,
		deployment: &k8sDeployment,
	}
	Convey("Test Container type", t, func() {
		So(container.GetName(), ShouldEqual, "my-container")

		Convey("Test image semver", func() {
			version, err := container.ParseImageVersion()
			So(err, ShouldBeNil)
			So(version.Major, ShouldEqual, 1)
			So(version.Minor, ShouldEqual, 2)
			So(version.Patch, ShouldEqual, 3)

			container.container.Image = "registry.example.com/my-image:latest"
			version, err = container.ParseImageVersion()
			So(err, ShouldNotBeNil)
			So(version.Major, ShouldEqual, 0)
			So(version.Minor, ShouldEqual, 0)
			So(version.Patch, ShouldEqual, 0)

			container.container.Image = "registry.example.com/my-image"
			version, err = container.ParseImageVersion()
			So(err, ShouldNotBeNil)
			So(version.Major, ShouldEqual, 0)
			So(version.Minor, ShouldEqual, 0)
			So(version.Patch, ShouldEqual, 0)
		})

		Convey("Test update version", func() {
			container.container.Image = "registry.example.com/my-image:1.2.3"

			newVersion, err := semver.Make("1.6.6")
			So(err, ShouldBeNil)

			newContainer, err := container.UpdateImageVersion(newVersion)
			So(err, ShouldBeNil)
			So(newContainer.container.Image, ShouldEqual, "registry.example.com/my-image:1.6.6")
		})
	})
}

// Uncomment for some integration testing

// func TestKube(t *testing.T) {
// Convey("Test update deployment", t, func() {
// k, err := util.CreateClient("http://localhost:8001")
// So(err, ShouldBeNil)

// list, err := NewList(k, api.NamespaceDefault)
// So(err, ShouldBeNil)

// ver, err := semver.Make("1.2.3")
// So(err, ShouldBeNil)

// err = list.Items[0].UpdateDeployment(k, api.NamespaceDefault, ver)
// So(err, ShouldBeNil)
// })
// }
