echo $TRAVIS_REPO_SLUG

msg="{\"msgtype\": \"text\", \"text\": { \"content\": \"failure, reason: $TRAVIS_REPO_SLUG\" }}"
curl 'https://oapi.dingtalk.com/robot/send?access_token=5349ab8e8a4ec739e5f2c704ebdcf50bf1e52c5b925d176a871c4e5b98735a20' \
	-H 'Content-Type: application/json' \
	-d "$msg"
