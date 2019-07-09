#!/usr/bin/env bash
export DING_TOKEN="5349ab8e8a4ec739e5f2c704ebdcf50bf1e52c5b925d176a871c4e5b98735a20"
if [ "${TRAVIS_REPO_SLUG}" != "logancloud/logan-app-operator" ]; then
    now=$(date -u +%s%N)
    elapsed_nano=`expr $now - $TRAVIS_SD_START_TIME`
    elapsed_time=`expr $elapsed_nano / 1000000000`
    error_info="Build #$TRAVIS_BUILD_NUMBER($TRAVIS_COMMIT) of $TRAVIS_REPO_SLUG@$TRAVIS_BRANCH $TRAVIS_SD_STATUS in $elapsed_time seconds"
    curl "https://oapi.dingtalk.com/robot/send?access_token=$DING_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{\"msgtype\": \"text\", \"text\": { \"content\": \"$TRAVIS_SD_STATUS, reason: $error_info\" }}"
fi
