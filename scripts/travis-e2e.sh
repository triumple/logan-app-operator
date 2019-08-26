#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u
# print each command before executing it
set -x

SCRIPT_DIR=$(dirname "${BASH_SOURCE[0]}")

"${SCRIPT_DIR}"/create-minikube.sh

#init project logan
kubectl create namespace logan
oc project logan

# e2e images
export REPO="logancloud/logan-app-operator"
docker tag ${REPO}:latest "${REPO}:latest-e2e"

#init operator
make initdeploy
oc replace -f test/resources/operator-e2e.yaml
oc scale deploy logan-app-operator --replicas=1
JSONPATH='{range .items[*]}{@.metadata.name}:{range @.status.conditions[*]}{@.type}={@.status};{end}{end}'; until kubectl -n logan get pods -lname=logan-app-operator -o jsonpath="$JSONPATH" 2>&1 | grep -q "Ready=True"; do sleep 1;echo "waiting for logan-app-operator to be available"; kubectl get pods --all-namespaces; done
oc replace configmap --filename test/resources/config.yaml

#run test
ginkgo -p --skip="\[Serial\]" -r test
ginkgo --focus="\[Serial\]" -r test

#"${SCRIPT_DIR}"/delete-minikube.sh