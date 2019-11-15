#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u
# print each command before executing it
set -x

# prerequisite check
set +e
set +x
result="ok"
go version >/dev/null 2>&1
if [ $? != 0 ]; then
    echo "ERROR: please install go"
    result="error"
fi

operator-sdk version >/dev/null 2>&1
if [ $? != 0 ]; then
    echo "ERROR: please install operator-sdk"
    result="error"
fi

ginkgo version >/dev/null 2>&1
if [ $? != 0 ]; then
    echo "ERROR: please install ginkgo"
    result="error"
fi

kubectl >/dev/null 2>&1
if [ $? != 0 ]; then
    echo "ERROR: please install kubectl"
    result="error"
fi

minikube version >/dev/null 2>&1
if [ $? != 0 ]; then
    echo "ERROR: please install minikube"
    result="error"
fi

oc >/dev/null 2>&1
if [ $? != 0 ]; then
    echo "ERROR: please install oc"
    result="error"
fi

if [ ${result} == "error" ]; then
    echo "ERROR: prerequisite check fail"
    exit 1
fi
set -x
set -e

export MINIKUBE_VERSION=v1.2.0
export KUBERNETES_VERSION=v1.12.0

export MINIKUBE_HOME=$HOME
export CHANGE_MINIKUBE_NONE_USER=true
if [ ! -d "${HOME}/.kube" ];then
    mkdir "${HOME}"/.kube || true
fi
touch "${HOME}"/.kube/config

export KUBECONFIG=$HOME/.kube/config

sudoCmd=""
if [ "$(id -u)" != "0" ]; then
    sudoCmd="sudo"
fi

# minikube config
minikube config set WantUpdateNotification false
minikube config set WantReportErrorPrompt false
minikube config set WantNoneDriverWarning false

minikube version

registry=""
profile=""
# check skip test
set +u
    if [ $(uname) == "Darwin" ]; then
        # set insecure-registry ip on mac host, default 192.168.99.0/24
        # can use env variable INSECURE_REGISTRY to set insecure-registry ip
        if [ "${INSECURE_REGISTRY}x" != "x" ]; then
            registry="--insecure-registry ${INSECURE_REGISTRY}"
        else
            registry="--insecure-registry 192.168.99.0/24"
        fi

        # on mac host, should use virtualbox, for details: https://minikube.sigs.k8s.io/docs/start/macos/
        minikube config set vm-driver virtualbox
    elif [ $(uname) == "Linux" ]; then
        # on linux host, should set vm-driver none, for details: https://minikube.sigs.k8s.io/docs/start/linux/
        minikube config set vm-driver none
    fi

    # use profile to label the minikube used for e2e local testing
    if [ "${1}x" == "localx" ]; then
        profile="--profile e2e-local"
    fi
set -u


${sudoCmd} minikube start --kubernetes-version=$KUBERNETES_VERSION --extra-config=apiserver.authorization-mode=RBAC ${registry} ${profile}

# enable registry to store image on mac virtualbox
if [ $(uname) == "Darwin" ]; then
    ${sudoCmd} minikube addons enable registry ${profile}
fi

set +u
if [ "${1}x" != "localx" ]; then
    ${sudoCmd} chown -R travis: /home/travis/.minikube/
else
    ${sudoCmd} chown -R $USER: $HOME/.minikube/
fi
set -u

${sudoCmd} minikube update-context ${profile}

# waiting for node(s) to be ready
JSONPATH='{range .items[*]}{@.metadata.name}:{range @.status.conditions[*]}{@.type}={@.status};{end}{end}'; until kubectl get nodes -o jsonpath="$JSONPATH" 2>&1 | grep -q "Ready=True"; do sleep 1; done

# waiting for kube-addon-manager to be ready
JSONPATH='{range .items[*]}{@.metadata.name}:{range @.status.conditions[*]}{@.type}={@.status};{end}{end}'; until kubectl -n kube-system get pods -lcomponent=kube-addon-manager -o jsonpath="$JSONPATH" 2>&1 | grep -q "Ready=True"; do sleep 1;echo "waiting for kube-addon-manager to be available"; kubectl get pods --all-namespaces; done

# waiting for kube-dns to be ready
JSONPATH='{range .items[*]}{@.metadata.name}:{range @.status.conditions[*]}{@.type}={@.status};{end}{end}'; until kubectl -n kube-system get pods -lk8s-app=kube-dns -o jsonpath="$JSONPATH" 2>&1 | grep -q "Ready=True"; do sleep 1;echo "waiting for kube-dns to be available"; kubectl get pods --all-namespaces; done

if [ $(uname) == "Darwin" ]; then
    # waiting for registry-proxy to be ready
    JSONPATH='{range .items[*]}{@.metadata.name}:{range @.status.conditions[*]}{@.type}={@.status};{end}{end}'; until kubectl -n kube-system get pods -lkubernetes.io/minikube-addons=registry -o jsonpath="$JSONPATH" 2>&1 | grep -o "Ready=True" | if [ $(grep -c "Ready=True") == 4 ]; then true; else false; fi; do sleep 1;echo "waiting for registry/registry-proxy to be available"; kubectl get pods --all-namespaces; done
fi

kubectl apply -f scripts/minikube-rbac.yaml