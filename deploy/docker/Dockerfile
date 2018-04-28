FROM alpine
LABEL maintainers="Kubernetes Authors"
LABEL description="LVM CSI Plugin"

RUN apk update && apk add blkid file util-linux e2fsprogs
COPY lvmplugin /lvmplugin

ENTRYPOINT ["/lvmplugin"]
