echo $TRAVIS_REPO_SLUG
elapsed_nano=`expr $(date -u +%s%N) - $TRAVIS_TIMER_START_TIME`
elapsed_time=`expr $elapsed_nano / 1000000000`
error_info="Build #$TRAVIS_BUILD_NUMBER($TRAVIS_COMMIT) of $TRAVIS_REPO_SLUG@$TRAVIS_BRANCH fail in $elapsed_time seconds"
msg="{\"msgtype\": \"text\", \"text\": { \"content\": \"failure, reason: $error_info\" }}"
echo $msg
curl 'https://oapi.dingtalk.com/robot/send?access_token=5349ab8e8a4ec739e5f2c704ebdcf50bf1e52c5b925d176a871c4e5b98735a20' \
	-H 'Content-Type: application/json' \
	-d "$msg"
