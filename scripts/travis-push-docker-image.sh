#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u

export REPO="logancloud/logan-app-operator"
echo "1"
if [[ "${TRAVIS_PULL_REQUEST}" = "false" ]]; then
    docker login -u "$DOCKER_USERNAME" -p "$DOCKER_PASSWORD"
    docker push ${REPO}:latest
fi
echo "2"
pullRequstId=""
echo "21"
gitlogs="$(git log -1 | grep \"Merge pull request\")"
echo "22"
gitlogs="Merge pull request #44 from triumple/operator_timezone_fix"
re="Merge pull request #([0-9]+) .*"
if [[ $gitlogs =~ $re ]]; then
    echo "3"
    echo ${BASH_REMATCH[1]}
    pullRequstId=${BASH_REMATCH[1]}
    echo "4"
fi

echo $pullRequstId
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
