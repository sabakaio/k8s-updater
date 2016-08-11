FROM alpine:3.4

RUN apk add --update ca-certificates \
    && mkdir -p /etc/ssl/certs/ \
    && update-ca-certificates --fresh \
    && rm -rf /var/cache/apk/* /tmp/*

COPY k8s-updater /usr/local/bin/
CMD ["k8s-updater"]

