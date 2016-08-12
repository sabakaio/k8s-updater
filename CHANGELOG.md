# k8s-updater changelog

## 0.0.2

### bugfixes

- all containers in a list to update had a link to the one *Deployment* because of variable passed by link

### features

- configure `loglevel`, *Kubernetes* api `host` and `namespace` using [viper](https://github.com/spf13/viper)
- dry run mode with `--dry-run` flag or `APP_DRYRUN` environment variable
