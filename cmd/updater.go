package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/sabakaio/k8s-updater/pkg/updater"
	"github.com/spf13/viper"
)

func update() {
	list, err := updater.NewList(k, viper.GetString("namespace"))
	if err != nil {
		log.Fatalln("Can't get deployments", err)
	}
	if len(list.Items) == 0 {
		log.Warningln("No autoupdate deployments found")
	}
	for _, c := range list.Items {
		newVersion, err := c.GetAutoupdateVersion()
		if err != nil {
			log.Errorln(err)
		}
		if newVersion != nil {
			log.Infof("deployment=%s container=%s giong to update up to version %s", c.GetDeploymentName(), c.GetName(), newVersion.String())
			c.UpdateDeployment(k, *newVersion)
		} else {
			log.Debugf("deployment=%s container=%s nothing to update", c.GetDeploymentName(), c.GetName())
		}
	}
}
