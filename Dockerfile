FROM registry.access.redhat.com/ubi9/go-toolset:latest

COPY . /src
WORKDIR /src
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -a  -ldflags="-s -w"  -installsuffix cgo -o /tmp/go-simple-uploader .

FROM registry.access.redhat.com/ubi9/ubi-minimal:latest
RUN microdnf -y update && microdnf -y --nodocs install tar rsync shadow-utils && microdnf clean all && rm -rf /var/cache/yum

COPY --from=0 /tmp/go-simple-uploader /usr/local/bin/go-simple-uploader

USER 1001
EXPOSE 8080
CMD ["/usr/local/bin/go-simple-uploader"]
