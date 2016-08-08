FROM alpine:3.4
COPY k8s-updater /usr/local/bin/
CMD ["k8s-updater"]

