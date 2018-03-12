FROM ubuntu:18.04

MAINTAINER Ryan Gerstenkorn version: 0.1

WORKDIR /go/src/github.com/RyanJarv/dockersnitch 
ENV GOPATH /go
ENV PATH /go/bin/:${PATH}

RUN apt-get update 
RUN apt-get install -y libnetfilter-queue-dev libnetfilter-queue1
RUN apt-get install -y iptables ipset
RUN apt-get install -y golang sudo
RUN apt-get clean && rm -rf /var/lib/apt/lists/*

COPY main.go /go/src/github.com/RyanJarv/dockersnitch 
COPY dockersnitch /go/src/github.com/RyanJarv/dockersnitch/dockersnitch
COPY vendor /go/src/github.com/RyanJarv/dockersnitch/vendor

RUN mkdir -p /go/bin/; go build -o /go/bin/dockersnitch
RUN rm -rf /go/src

CMD ["/go/bin/dockersnitch"]
