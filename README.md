[![GolangCI](https://golangci.com/badges/github.com/chmouel/go-simple-uploader.svg)](https://golangci.com/r/github.com/chmouel/go-simple-uploader)
[![License](https://img.shields.io/github/license/chmouel/go-simple-uploader)](/LICENSE)

# GO Simple Uploader

A simple uploader in GO, meant to be deployed in a container behind a
nginx protected environment, but you can deploy it as you wish.

I mainly deploy it to OpenShift which handles for me the building of the source
the deployment with a nginx server and observability etc See the OpenShift
deployment section of this document.

## Install

```shell
go install github.com/chmouel/go-simple-uploader
```

### Configuration

done by environment variable with :

- **UPLOADER_HOST** -- hostname to bind to
- **UPLOADER_PORT** -- port to bind to
- **UPLOADER_DIRECTORY** -- Directory where to upload
- **UPLOADER_UPLOAD_CREDENTIALS** -- If you like to protect your upload directory by a username password then specify them separated by colon, i.e: `username:password`

Usually you will run this behind an HTTP server which will handle the upload
protection, acl or others. You really don't want to expose this to the internet
unprotected.

## Usage

It accepts form http fields:

- **file**: The file stream of the upload
- **path**: The path
- **targz**: Assume the file uploaded is a tarball which we want to uncompress on filesystem

## OpenShift Deployment

This uses S2I to generate an image against the repo and output it to an
OpenShift ImageStream and then use a Kubernetes `Deployment` to deploy it, with
a few sed for dynamic variables.

The deployment has two containers, the main one is nginx getting all requests and
passing the uploads to the uwsgi process in the other container and serves the
static file directly.

Under nginx configuration the `/private` directory is protected with the same
username password as configured in the htpasswd, which you can use to 'protect
stuff.

## Setup

## Run directly

You can run the service directly with the kubernetes [template](kubernetes/deployment.yaml). 

By default the uplod password and username is username:password, to protect the `/upload` and `/delete` (as you should) properly you will need to change the secret from https://github.com/chmouel/go-simple-uploader/blob/master/kubernetes/deployment.yaml#L5

## Run behind nginx

You need first to create a username password with :

```shell
htpasswd -b -c openshift/config/osinstall.htpasswd username password
```

Then you just use the makefile target to build and deploy :

```shell
oc new-project uploader && make deploy
```

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

### API

#### Upload**

- **method**: POST
- **path**: */upload*
- **arguments**:
- **path**: Path where to upload the files, which is relative to the upload directory, directory traversal is checked and disallowed.
- **file**: File post data
- **targz**: Booleean if we want to uncompress the file on fs

- **examples**:

```shell
curl -u username:password -F path=hello-upload.txt -X POST -F file=@/tmp/hello.txt ${route}/upload
```

```shell
tar czf - /path/to/directory|curl -u username:password -F path=hello-upload.txt -F targz=true -X POST -F file=@- ${route}/upload
```

### Delete

- **method**: DELETE

- **path**: */upload*
- **arguments**:
- **path**: Path to delete

- **example**:

```shell
curl -u username:password -F path=hello-upload.txt -X DELETE ${route}/upload
```
---
- **method**: DELETE

- **path**: */delete*
- **arguments**:
- **path**: path to directory to delete files in it
- **days**: delete files in above directory older than X `days` 
- **recursive**: flag to recursively delete files child directorires of `path` (defaults to `false`).  

- **example**:

```shell
curl -k -s -u username:password -F path=/path/to/directory  -F days=1 -F recursive=true  -X DELETE ${route}/delete
```

## [LICENSE](LICENSE)

[Apache 2.0](LICENSE)
