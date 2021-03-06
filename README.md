[![Build Status](https://travis-ci.org/sabakaio/k8s-updater.svg?branch=master)](https://travis-ci.org/sabakaio/k8s-updater)

# k8s-updater

**Caution, this project is in alpha stage.**

## Purpose

Provide container updates to latest image versions at *Kubernetes* cluster.

## Usage 

Deploy this image as a job on your cluster. You can run it manually or schedule with [sabakaio/kron](https://github.com/sabakaio/kron) (as in following example)

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: autoupdater
  labels:
    kron: "true" # Schedure with the sabakaio/kron
  annotations:
    schedule: "@every 1h" # Run every hour with sabakaio/kron
spec:
  template:
    spec:
      containers:
        - image: sabaka/k8s-updater:0.0.1
          name: autoupdater
      restartPolicy: Never # Important for containers that should run just once (by schedule)
```

Updater is going to list all *Deployments* labled with `autoupdate` to perform version check and updates

```yaml
metadata:
  labels:
    autoupdate: "true"
```

To perform some pre-update actions (e.g. run database migrations) you can setup a hook for each container with *Deployment* annotations

```yaml
metadata:
  annotations:
    before_autoupdate_web: "migration" # The job to run before autoupdate container `web`
```

You could also limit a versions range to to updrade on, with *Deployment* annotations

```yaml
metadata:
  annotations:
    autoupdate_version_range_web: ">1.0.0 <2.0.0"
```
