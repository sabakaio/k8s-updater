package cmd

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/sabakaio/k8s-updater/pkg/updater"
)

func update() {
	list, err := updater.NewList(k, namespace)
	if err != nil {
		log.Fatalln("Can't get deployments", err)
	}
	if len(list.Items) == 0 {
		log.Warningln("No autoupdate deployments found")
	}
	for _, c := range list.Items {
		version, err := c.GetImageVersion()
		if err != nil {
			msg := fmt.Sprintf("could not get container image version for '%s'", c.GetName())
			log.Warningln(msg, err)
		}
		latest, err := c.GetLatestVersion()
		if err != nil {
			log.Error(err)
		}
		msg := fmt.Sprintf("'%s' container of '%s' cluster: current version is %s, latest is %s",
			c.GetName(), c.GetDeploymentName(), version.String(), latest.String())
		// Update container deployment if greater image version found
		if latest.Semver.GT(version.Semver) {
			log.Infof("going to update %s", msg)
			c.UpdateDeployment(k, *latest)
		} else {
			log.Debugf("nothing to update for %s", msg)
		}
	}
}
