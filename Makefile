.PHONY: test

repo = github.com/sabakaio/k8s-updater

test:
	go test -v $(repo)/pkg/updater
