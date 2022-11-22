FROM registry.access.redhat.com/ubi9/go-toolset:latest
COPY . /go/src/github.com/chmouel/go-simple-uploader
WORKDIR /go/src/github.com/chmouel/go-simple-uploader
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o go-simple-uploader .

FROM registry.access.redhat.com/ubi9/ubi-minimal:latest
COPY --from=0 /go/src/github.com/chmouel/go-simple-uploader/go-simple-uploader /usr/local/bin/go-simple-uploader

WORKDIR /
EXPOSE 8080
CMD ["/usr/local/bin/go-simple-uploader"]
