docker build -t registry.logan.xiaopeng.local/logancloud/operator-e2e:latest .
docker push registry.logan.xiaopeng.local/logancloud/operator-e2e:latest



oc adm policy add-scc-to-user privileged -n logan -z default

adduser travis
usermod -aG sudo travis

sudo visudo
travis  ALL = NOPASSWD: /usr/local/bin/minikube


docker images docker| less
docker run -it --privileged --name some-docker -d docker:dind
docker logs some-docker
docker run -it --rm --link some-docker:docker docker version
docker run -it --rm --link some-docker:docker docker sh
  docker version
docker run -it --rm --link some-docker:docker docker info
docker rm -f some-docker
docker run -it --privileged --name some-docker -d docker:dind --storage-driver=devicemapper



docker run -it --privileged --name some-docker -d docker:18-dind



docker run -it --privileged --name some-docker1 -d registry.logan.xiaopeng.local/logancloud/operator-dind:latest



docker exec -it c38765334e33 /bin/sh
export DOCKER_HOST=tcp://172.17.0.3:2375 //ip地址为dind的地址

docker run -it --rm --link some-docker:docker ubuntu bash


docker run -it --privileged --name some-docker --mount type=bind,source=/etc/docker/daemon.json,target=/etc/docker/daemon.json -d docker:dind