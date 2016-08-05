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
	})
}
