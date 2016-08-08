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
		version, err := c.ParseImageVersion()
		if err != nil {
			msg := fmt.Sprintf("could not parse container image version for '%s'", c.GetName())
			log.Warningln(msg, err)
		}
		latest, err := c.GetLatestVersion()
		if err != nil {
			log.Error(err)
		}
		log.Debugln(c.GetName(), version.String(), "=>", latest.String())
	}
}
