package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/sabakaio/k8s-updater/pkg/util"
)

func update() {
	list, err := util.ListDeployments(k, namespace)
	if err != nil {
		log.Fatalln("Can't get deployments", err)
	}
	if len(list.Items) == 0 {
		log.Warningln("No autoupdate deployments found")
	}
	for _, d := range list.Items {
		log.Debugln(d.GetName())
	}
}
