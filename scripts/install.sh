#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u
# print each command before executing it
set -x

export MINIKUBE_VERSION=v1.2.0
export KUBERNETES_VERSION=v1.12.0

if [ -f $GOPATH/bin/operator-sdk ];then
    echo "operator-sdk in cached install"
else
    wget https://github.com/operator-framework/operator-sdk/releases/download/v0.10.1/operator-sdk-v0.10.1-x86_64-linux-gnu
    mv operator-sdk-v0.10.1-x86_64-linux-gnu $GOPATH/bin/operator-sdk
    chmod +x $GOPATH/bin/operator-sdk
fi

if [ -f $GOPATH/bin/ginkgo ];then
    echo "ginkgo in cached install"
else
    go get -u github.com/onsi/ginkgo/ginkgo
fi

sudo mount --make-rshared /
sudo mount --make-rshared /proc
sudo mount --make-rshared /sys

if [ -f $GOPATH/bin/kubectl ];then
    echo "kubectl in cached install"
    sudo cp $GOPATH/bin/kubectl /usr/local/bin/
else
    curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/$KUBERNETES_VERSION/bin/linux/amd64/kubectl && \
        chmod +x kubectl &&  \
        sudo mv kubectl /usr/local/bin/
        sudo cp /usr/local/bin/kubectl $GOPATH/bin/
fi

if [ -f $GOPATH/bin/minikube ];then
    echo "minikube in cached install"
    sudo cp $GOPATH/bin/minikube /usr/local/bin/
else
    curl -Lo minikube https://storage.googleapis.com/minikube/releases/$MINIKUBE_VERSION/minikube-linux-amd64 && \
        chmod +x minikube && \
        sudo mv minikube /usr/local/bin/
        sudo cp /usr/local/bin/minikube $GOPATH/bin/
fi

if [ -f $GOPATH/bin/oc ];then
    echo "oc in cached install"
    sudo cp $GOPATH/bin/oc /usr/local/bin/
else
    curl -Lo /tmp/oc.tar.gz https://github.com/openshift/origin/releases/download/v3.11.0/openshift-origin-client-tools-v3.11.0-0cbc58b-linux-64bit.tar.gz  && \
        tar -zxvf /tmp/oc.tar.gz -C /tmp &&  \
        sudo mv /tmp/openshift-origin-client-tools-v3.11.0-0cbc58b-linux-64bit/oc /usr/local/bin/
        sudo cp /usr/local/bin/oc $GOPATH/bin/
fi