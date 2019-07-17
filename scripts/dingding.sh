#!/usr/bin/env  bash

if [[ "${DING_TOKEN}" != "" ]]; then
    now=$(date -u +%s)
    elapsed_seconds=`expr $now - $TRAVIS_SD_START_TIME`

    author=""
    git_logs=$(git log -1 | grep "Author: ")
    re="Author: (.*) <.*"
    if [[ $git_logs =~ $re ]]; then
        author=${BASH_REMATCH[1]}
    fi

    output_info="Build [#$TRAVIS_BUILD_NUMBER]($TRAVIS_BUILD_WEB_URL) of $TRAVIS_REPO_SLUG@$TRAVIS_BRANCH by $author $TRAVIS_SD_STATUS in $elapsed_seconds seconds"
    curl "https://oapi.dingtalk.com/robot/send?access_token=$DING_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{\"msgtype\": \"markdown\", \"markdown\": { \"title\": \"logan-app-operator travis info\", \"text\": \"$TRAVIS_SD_STATUS to build logan-app-operator, info: $output_info\" }}"
fi