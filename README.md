# GO Simple Uploader

A simple uploader in GO, is is meant to be deployed in a container behind a
nginx deployment but you can deploy it as you wish.

I mainly deploy it to an OpenShift.

## Build

```shell
go build github.com/chmouel/go-simple-uploader
```

### Configuration

done by environment variable with :

**UPLOADER_HOST** -- hostname to bind to
**UPLOADER_PORT** -- port to bind to
**UPLOADER_DIRECTORY** -- Directory where to upload

## Usage

It accepts two form field  :

**file**: The file stream of the upload
**path**: The path

## OpenShift Deployment

This uses S2I to generate an image against the repo and output it to an
OpenShift ImageStream and then use a Kubernetes `Deployment` to deploy it, with
a few sed for dynamic variables.

The deployment has two containers, the main one is nginx getting all request and
passing the uploads to the uwsgi process in the other container and serves the
static file directly.

## Install

You need first to create a a username password with :

```
htpasswd -b -c openshift/config/osinstall.htpasswd username password
```

Then you just use the makefile target, build and deploy you just need :

```
oc new-project uploader && make deploy
```

## Usage

Get the route of your deployment with :

```shell
oc get route uploader -o jsonpath='{.spec.host}'`
```

Test as working with :

```shell
route=http://$(oc get route uploader -o jsonpath='{.spec.host}')
echo "HELLO WORLD" > /tmp/hello.txt
curl -u username:password -F path=hello-upload.txt -X POST -F file=@/tmp/hello.txt ${route}/upload
curl ${route}/hello-upload.txt
```
