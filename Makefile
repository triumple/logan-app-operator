# Image URL to use all building/pushing image targets
IMG ?= logancloud/logan-app-operator:latest

all: test

# Run tests
test:
	ginkgo -r -p

# Run against the configured Kubernetes cluster in ~/.kube/config
run: fmt vet
	operator-sdk up local --namespace=logan --operator-flags "--config=configs/config_local.yaml --zap-devel --zap-level info"

rundebug: fmt vet
	operator-sdk up local --namespace=logan --operator-flags "--config=configs/config_local.yaml --zap-devel"

rundev:
	LOGAN_ENV=dev WATCH_NAMESPACE=logan-dev operator-sdk up local --namespace=logan-dev --operator-flags "--config=configs/config_local.yaml --zap-devel"

runprod:
	LOGAN_ENV=prod WATCH_NAMESPACE=logan-prod operator-sdk up local --namespace=logan-prod --operator-flags "--config=configs/config_local.yaml --zap-devel"

# Install CRDs into a cluster
install:
	kubectl apply -f deploy

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy:
	kubectl apply -f deploy/crds

# Run go fmt against code
fmt:
	go fmt ./pkg/... ./cmd/...

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...

# Run generate k8s
gen-k8s:
	operator-sdk generate k8s

# Build
build: docker-build docker-push

# Build the docker image
docker-build:
	operator-sdk build ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

travis-build:
	./scripts/travis-push-docker-image.sh

# Redploy configmap
recm:
	oc delete configmap logan-app-operator-config --ignore-not-found=true
	oc create configmap logan-app-operator-config --from-file=configs/config.yaml

rerole:
	oc delete -f deploy/role.yaml --ignore-not-found=true
	oc create -f deploy/role.yaml
	oc delete -f deploy/role_binding.yaml --ignore-not-found=true
	oc create -f deploy/role_binding.yaml
	oc delete -f deploy/role_operator.yaml --ignore-not-found=true
	oc create -f deploy/role_operator.yaml
	oc delete -f deploy/service_account.yaml --ignore-not-found=true
	oc create -f deploy/service_account.yaml

recrd:
	oc delete -f deploy/crds/app_v1_javaboot_crd.yaml --ignore-not-found=true
	oc create -f deploy/crds/app_v1_javaboot_crd.yaml

	oc delete -f deploy/crds/app_v1_phpboot_crd.yaml --ignore-not-found=true
	oc create -f deploy/crds/app_v1_phpboot_crd.yaml

	oc delete -f deploy/crds/app_v1_pythonboot_crd.yaml --ignore-not-found=true
	oc create -f deploy/crds/app_v1_pythonboot_crd.yaml

	oc delete -f deploy/crds/app_v1_nodejsboot_crd.yaml --ignore-not-found=true
	oc create -f deploy/crds/app_v1_nodejsboot_crd.yaml

	oc delete -f deploy/crds/app_v1_webboot_crd.yaml --ignore-not-found=true
	oc create -f deploy/crds/app_v1_webboot_crd.yaml

# Redeploy controller in the configured Kubernetes cluster in ~/.kube/config
redeploy: recm
	oc delete -f deploy/operator-test.yaml -f deploy/operator-dev.yaml -n logan --ignore-not-found=true
	oc create -f deploy/operator-test.yaml -f deploy/operator-dev.yaml -n logan

# test java
test-java:
	oc delete -f examples/test-java.yaml --ignore-not-found=true
	oc create -f examples/test-java.yaml

# test php
test-php:
	oc delete -f examples/test-php.yaml --ignore-not-found=true
	oc create -f examples/test-php.yaml

# test python
test-python:
	oc delete -f examples/test-python.yaml --ignore-not-found=true
	oc create -f examples/test-python.yaml

# test nodejs
test-nodejs:
	oc delete -f examples/test-nodejs.yaml --ignore-not-found=true
	oc create -f examples/test-nodejs.yaml

# test web
test-web:
	oc delete -f examples/test-web.yaml --ignore-not-found=true
	oc create -f examples/test-web.yaml

test-all: test-java test-php test-python test-nodejs test-web

test-deleteall:
	oc delete -f examples/test-java.yaml --ignore-not-found=true
	oc delete -f examples/test-php.yaml --ignore-not-found=true
	oc delete -f examples/test-python.yaml --ignore-not-found=true
	oc delete -f examples/test-nodejs.yaml --ignore-not-found=true
	oc delete -f examples/test-web.yaml --ignore-not-found=true

test-createall:
	oc create -f examples/crds/test_java.yaml
	oc create -f examples/crds/test_php.yaml
	oc create -f examples/crds/test_python.yaml
	oc create -f examples/crds/test_node.yaml
	oc create -f examples/crds/test_web.yaml

#  test recreate 100 times
test-batch:
	scripts/all.sh

