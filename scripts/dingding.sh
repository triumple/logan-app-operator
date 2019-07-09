#!/usr/bin/env bash

if [ "${DING_TOKEN}" != "" ]; then
    now=$(date -u +%s)
    elapsed_seconds=`expr $now - $TRAVIS_SD_START_TIME`
    output_info="Build #$TRAVIS_BUILD_NUMBER($TRAVIS_COMMIT) of $TRAVIS_REPO_SLUG@$TRAVIS_BRANCH $TRAVIS_SD_STATUS in $elapsed_seconds seconds"
    curl "https://oapi.dingtalk.com/robot/send?access_token=$DING_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{\"msgtype\": \"text\", \"text\": { \"content\": \"$TRAVIS_SD_STATUS to build logan-app-operator, info: $output_info\" }}"
fi
