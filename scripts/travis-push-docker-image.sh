#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u

export REPO="logancloud/logan-app-operator"
if [[ "${TRAVIS_PULL_REQUEST}" = "false" ]]; then
    docker login -u "$DOCKER_USERNAME" -p "$DOCKER_PASSWORD"
    docker push ${REPO}:latest
fi

pullRequstId=""
gitlogs='$(git log -1 | grep "Merge pull request")'
#gitlogs="Merge pull request #44 from triumple/operator_timezone_fix"
re="Merge pull request #([0-9]+) .*"
if [[ $gitlogs =~ $re ]]; then
    pullRequstId=${BASH_REMATCH[1]}
    echo "4"
fi

echo "test_$pullRequstId"
if [[ "${pullRequstId}" != "" ]]; then
    export TAG="pr_${pullRequstId}"
    docker tag ${REPO}:latest "${REPO}:${TAG}"
    docker images
    docker login -u "$DOCKER_USERNAME" -p "$DOCKER_PASSWORD"
    echo "Pushing to docker hub ${REPO}:${TAG}"
    docker push "${REPO}:${TAG}"
fi

if [[ "${TRAVIS_TAG}" != "" ]]; then
	# For both git tags and git branches 'TRAVIS_BRANCH' contains the name.
    export TAG="${TRAVIS_BRANCH}"
    docker tag ${REPO}:latest "${REPO}:${TAG}"
    docker login -u "$DOCKER_USERNAME" -p "$DOCKER_PASSWORD"
    echo "Pushing to docker hub ${REPO}:${TAG}"
    docker push "${REPO}:${TAG}"
fi
