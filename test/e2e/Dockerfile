FROM ubuntu:16.04

RUN apt-get update
RUN apt-get install -y curl wget docker.io sudo vim systemd
RUN rm -rf /var/lib/apt/lists/*

ENV GOLANG_VERSION 1.12.5

RUN curl -sSL https://storage.googleapis.com/golang/go$GOLANG_VERSION.linux-amd64.tar.gz \
		| tar -v -C /usr/local -xz

ENV PATH /usr/local/go/bin:$PATH

RUN mkdir -p /go/src /go/bin && chmod -R 777 /go
ENV GOROOT /usr/local/go
ENV GOPATH /go
ENV PATH /go/bin:$PATH
ENV DOCKER_HOST tcp://localhost:2375

RUN mkdir -p /test
COPY create-minikube.sh /test/.
COPY install.sh /test/.
COPY kubectl /test/.
COPY minikube /test/.
COPY oc.tar.gz /test/.
COPY operator-sdk-v0.8.1-x86_64-linux-gnu /test/.

WORKDIR /go

ADD run.sh /tmp/run.sh
RUN chmod +x /tmp/run.sh
CMD ["nohup", "sh", "/tmp/run.sh", "&"]
ENTRYPOINT ["nohup", "sh", "/tmp/run.sh", "&"]