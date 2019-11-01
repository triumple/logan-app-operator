#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u
# print each command before executing it
set -x

sudoCmd=""
if [ "$(id -u)" != "0" ]; then
    sudoCmd="sudo"
fi

export KUBECONFIG=$HOME/.kube/config

minikube version
${sudoCmd} minikube delete