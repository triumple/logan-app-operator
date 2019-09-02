#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u
# print each command before executing it
set -x



if [ -f $GOPATH/bin/operator-sdk ];then
    echo "operator-sdk in cached install"
else
    wget https://github.com/operator-framework/operator-sdk/releases/download/v0.8.1/operator-sdk-v0.8.1-x86_64-linux-gnu
    mv operator-sdk-v0.8.1-x86_64-linux-gnu $GOPATH/bin/operator-sdk
    chmod +x $GOPATH/bin/operator-sdk
fi

if [ -f $GOPATH/bin/dep ];then
    echo "dep in cached install"
else
    go get -u github.com/golang/dep/cmd/dep
fi

if [ -f $GOPATH/bin/ginkgo ];then
    echo "ginkgo in cached install"
else
    go get -u github.com/onsi/ginkgo/ginkgo
fi
