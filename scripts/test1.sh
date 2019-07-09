#!/usr/bin/env bash

pullRequstId=""
gitlogs="$(git log -1 | grep "Merge pull request")"
gitlogs="Merge pull request #44 from triumple/operator_timezone_fix"
re="Merge pull request #([0-9]+) .*"
if [[ $gitlogs =~ $re ]]; then
    pullRequstId=${BASH_REMATCH[1]};
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