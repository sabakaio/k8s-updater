FROM alpine:3.4
COPY kron /usr/local/bin/
CMD ["k8s-updater"]

