#!/usr/bin/env bash

if [ "${TRAVIS_REPO_SLUG}" != "logancloud/logan-app-operator" ]; then
    echo $TRAVIS_REPO_SLUG
    now=$(date -u +%s%N)
    elapsed_nano=`expr $now - $TRAVIS_SD_START_TIME`
    echo $now
    echo $TRAVIS_SD_START_TIME
    echo $elapsed_nano
    elapsed_time=`expr $elapsed_nano / 1000000000`
    error_info="Build #$TRAVIS_BUILD_NUMBER($TRAVIS_COMMIT) of $TRAVIS_REPO_SLUG@$TRAVIS_BRANCH $TRAVIS_SD_STATUS in $elapsed_time seconds"
    msg="{\"msgtype\": \"text\", \"text\": { \"content\": \"failure, reason: $error_info\" }}"
    echo $msg
    curl 'https://oapi.dingtalk.com/robot/send?access_token=5349ab8e8a4ec739e5f2c704ebdcf50bf1e52c5b925d176a871c4e5b98735a20' \
        -H 'Content-Type: application/json' \
        -d "$msg"
fi
