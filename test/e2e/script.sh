docker build -t registry.logan.xiaopeng.local/logancloud/operator-e2e:latest .
docker push registry.logan.xiaopeng.local/logancloud/operator-e2e:latest



oc adm policy add-scc-to-user privileged -n logan -z default