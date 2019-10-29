#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u
# print each command before executing it
set -x


# check skip test
set +u
    if [ "${SKIP_TEST}x" != "x" ]; then
        exit 0
    fi
set -u

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

oc replace -f test/resources/operator-e2e-dev.yaml
oc scale deploy logan-app-operator-dev --replicas=1
until kubectl -n logan get pods -lname=logan-app-operator-dev -o jsonpath="$JSONPATH" 2>&1 | grep -q "Ready=True"; do sleep 1;echo "waiting for logan-app-operator-dev to be available"; kubectl get pods --all-namespaces; done

oc replace configmap --filename test/resources/config.yaml

#run test
function runTest()
{
    set +e
    res="0"
    export GO111MODULE=on

    if [ $1 == "units" ]; then
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
    fi

    if [ $1 == "e2e" ]; then
        # run CONTROLLER test case
        ginkgo -p --focus="\[CONTROLLER\]" -r test
        sub_res=`echo $?`
        if [ $sub_res != "0" ]; then
            res=$sub_res
        fi
    fi

    if [ $1 == "integration" ]; then
        # run normal test case
        ginkgo --skip="\[Serial\]|\[Slow\]|\[Revision\]|\[CRD\]|\[CONTROLLER\]" -r test
        sub_res=`echo $?`
        if [ $sub_res != "0" ]; then
            res=$sub_res
        fi
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

    set -e
    exit $res
}
runTest $1

#"${SCRIPT_DIR}"/delete-minikube.sh