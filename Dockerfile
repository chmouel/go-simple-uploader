FROM mirror.gcr.io/library/golang:latest
COPY . /go/src/github.com/chmouel/go-simple-uploader
WORKDIR /go/src/github.com/chmouel/go-simple-uploader
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o go-simple-uploader .

FROM registry.access.redhat.com/ubi8/ubi-minimal:8.5-240.1648458092
COPY --from=0 /go/src/github.com/chmouel/go-simple-uploader/go-simple-uploader /usr/local/bin/go-simple-uploader

WORKDIR /
EXPOSE 8080
CMD ["/usr/local/bin/go-simple-uploader"]
