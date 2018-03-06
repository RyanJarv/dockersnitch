FROM ubuntu:18.04

MAINTAINER Ryan Gerstenkorn version: 0.1

RUN apt-get update && apt-get install -y libnetfilter-queue-dev libnetfilter-queue1 iptables golang && apt-get clean && rm -rf /var/lib/apt/lists/*


ENV GOPATH /go
WORKDIR /go/src/github.com/RyanJarv/dockersnitch

EXPOSE 52632

CMD ["/usr/bin/bash", "-l"]