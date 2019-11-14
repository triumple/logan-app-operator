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

profile=""
env=""
# check skip test
set +u
    if [ "${SKIP_TEST}x" != "x" ]; then
        exit 0
    fi

    if [ "${1}x" == "localx" ]; then
        set +e
        profile="--profile e2e-local"
        env=${1}
        set -e
    fi
set -u

SCRIPT_DIR=$(dirname "${BASH_SOURCE[0]}")

if [ $(${sudoCmd} minikube status ${profile} | grep -E Running\|Correctly\ Configured | wc -l) -ne 4 ]; then
    set +u
    if [ "${1}x" == "localx" ]; then
        rm -rf /etc/kubernetes
    fi
    set -u
    ${sudoCmd} chown -R travis: /home/travis/.minikube/
    "${SCRIPT_DIR}"/create-minikube.sh $env
fi

# delete project logan if existed
kubectl delete namespace logan --ignore-not-found=true
#init project logan
kubectl create namespace logan
oc project logan

# e2e images
if [ $(uname) == "Darwin" ]; then
    if [ $(docker ps -fname=socat_registry -fstatus=running | wc -l) -ne 2 ]; then
        docker run --name socat_registry -d --rm -it --network=host alpine ash -c "apk add socat && socat TCP-LISTEN:5000,reuseaddr,fork TCP:$(${sudoCmd} minikube ip ${profile}):5000"
    fi
    until docker ps -fname=socat_registry -fstatus=running | if [ $(wc -l)==2 ]; then true; else return false; fi; do sleep 1;echo "waiting for socat_registry to be available"; docker ps -fname=socat_registry -fstatus=running; done

    docker tag logancloud/logan-app-operator:latest localhost:5000/logancloud/logan-app-operator:latest-e2e
    until docker push localhost:5000/logancloud/logan-app-operator:latest-e2e; do sleep 1; echo "waiting for push image successfully"; done

    sed -i "" 's/image: logancloud\/logan-app-operator:latest-e2e/image: localhost:5000\/logancloud\/logan-app-operator:latest-e2e/g' test/resources/operator-e2e.yaml
    sed -i "" 's/image: logancloud\/logan-app-operator:latest-e2e/image: localhost:5000\/logancloud\/logan-app-operator:latest-e2e/g' test/resources/operator-e2e-dev.yaml
else
    export REPO="logancloud/logan-app-operator"
    docker tag ${REPO}:latest "${REPO}:latest-e2e"
fi

#init operator
make initdeploy
oc replace -f test/resources/operator-e2e.yaml
oc scale deploy logan-app-operator --replicas=1
JSONPATH='{range .items[*]}{@.metadata.name}:{range @.status.conditions[*]}{@.type}={@.status};{end}{end}'; until kubectl -n logan get pods -lname=logan-app-operator -o jsonpath="$JSONPATH" 2>&1 | grep -q "Ready=True"; do sleep 1;echo "waiting for logan-app-operator to be available"; kubectl get pods --all-namespaces; done

oc replace -f test/resources/operator-e2e-dev.yaml
oc scale deploy logan-app-operator-dev --replicas=1
until kubectl -n logan get pods -lname=logan-app-operator-dev -o jsonpath="$JSONPATH" 2>&1 | grep -q "Ready=True"; do sleep 1;echo "waiting for logan-app-operator-dev to be available"; kubectl get pods --all-namespaces; done

if [ $(uname) == "Darwin" ]; then
    docker stop socat_registry

    sed -i "" 's/image: localhost:5000\/logancloud\/logan-app-operator:latest-e2e/image: logancloud\/logan-app-operator:latest-e2e/g' test/resources/operator-e2e.yaml
    sed -i "" 's/image: localhost:5000\/logancloud\/logan-app-operator:latest-e2e/image: logancloud\/logan-app-operator:latest-e2e/g' test/resources/operator-e2e-dev.yaml
fi

oc replace configmap --filename test/resources/config.yaml

#run test
function runTest()
{
    set +e
    res="0"
    export GO111MODULE=on

    set +u
    if [ "${WAIT_TIME}x" == "x" ]; then
        export WAIT_TIME=1
    fi
    set -u

    # run revision test case
    ginkgo -p --focus="\[Revision\]" -r test
    sub_res=`echo $?`
    if [ $sub_res != "0" ]; then
        res=$sub_res
    fi

    # run serial test case
    ginkgo --focus="\[Serial\]" -r test
    sub_res=`echo $?`
    if [ $sub_res != "0" ]; then
        res=$sub_res
    fi

    # run CRD test case
    ginkgo -p --focus="\[CRD\]" -r test
    sub_res=`echo $?`
    if [ $sub_res != "0" ]; then
        res=$sub_res
    fi

    # run CONTROLLER test case
    ginkgo -p --focus="\[CONTROLLER\]" -r test
    sub_res=`echo $?`
    if [ $sub_res != "0" ]; then
        res=$sub_res
    fi

    # run normal test case
    ginkgo --skip="\[Serial\]|\[Slow\]|\[Revision\]|\[CRD\]|\[CONTROLLER\]" -r test
    sub_res=`echo $?`
    if [ $sub_res != "0" ]; then
        res=$sub_res
    fi


    # set WAIT_TIME, and run slow test case
    set +u
    if [ "${SLOW_WAIT_TIME}x" != "x" ]; then
        export WAIT_TIME=${SLOW_WAIT_TIME}
    else
        export WAIT_TIME=5
    fi
    set -u

    ginkgo -p --focus="\[Slow\]" -r test
    sub_res=`echo $?`
    if [ $sub_res != "0" ]; then
        res=$sub_res
    fi

    set +u
    if [ "${1}x" != "localx" ]; then
        set -e
    fi
    set -u

    if [ $res != "0" ]; then
        echo "ERROR: run e2e test case failed"
    fi

    return $res
}
runTest $env

#if [ $env == "local" ]; then
#   "${SCRIPT_DIR}"/delete-minikube.sh
#fi