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

echo "TRAVIS_COMMIT_MESSAGE=${TRAVIS_COMMIT_MESSAGE}"
echo "TRAVIS_COMMIT=${TRAVIS_COMMIT}"
echo "TRAVIS_COMMIT_RANGE=${TRAVIS_COMMIT_RANGE}"
if [[ "${TRAVIS_PULL_REQUEST}" != "false" ]]; then
    export TAG="pr_${TRAVIS_PULL_REQUEST}"
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
