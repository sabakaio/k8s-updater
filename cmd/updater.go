package cmd

import (
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
			log.Warningln("Could not parse container image version", err)
		}
		log.Debugln(c.GetName(), version)
	}
}
