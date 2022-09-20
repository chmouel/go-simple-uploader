SECRET_KEY := $(shell openssl rand -base64 20)
HTPASSWD_FILE := configs/osinstall.htpasswd
APP_NAME := uploader

ifeq ("$(wildcard $(HTPASSWD_FILE))","")
    $(error $(HTPASSWD_FILE) does not exist, create one with `htpasswd`)
endif

all:
	@echo "Use 'make deploy' to deploy, or 'make reconfig' to reconfig"

config:
	@oc delete configmap my-nginx-conf 2>/dev/null || true
	@oc create configmap --save-config my-nginx-conf --from-file $(HTPASSWD_FILE) \
		--from-file configs/nginx.conf

reconfig: config
	@oc delete pod `oc get pod -o name -l app=$(APP_NAME)|sed 's,pod/,,'` || true

deploy: reconfig
	@oc delete -f deployment.yaml 2>/dev/null || true
	@oc create -f deployment.yaml
	@oc expose service/$(APP_NAME) 2>/dev/null || true
	@echo "App is getting deployed to: "
	@echo http://`oc get route ${APP_NAME} -o jsonpath='{.spec.host}'`

.PHONY: deploy reconfig config all
