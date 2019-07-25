#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u
# print each command before executing it
set -x

go build -i -o $GOPATH/src/github.com/logancloud/logan-app-operator/build/_output/bin/logan-app-operator -gcflags all=-trimpath=${GOPATH} -asmflags all=-trimpath=${GOPATH} github.com/logancloud/logan-app-operator/cmd/manager

IMG=logancloud/logan-app-operator:latest
if [[ x$1 != x ]]
then
   IMG=$1
fi
docker build -f build/Dockerfile -t ${IMG} .