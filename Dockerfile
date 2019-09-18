FROM ubuntu:18.04

MAINTAINER Ryan Gerstenkorn version: 0.1

WORKDIR /app
ENV PATH /usr/local/go/bin:${PATH}
ENV GOVERSION 1.13

RUN apt-get update && apt-get dist-upgrade -y
RUN apt-get install -y libnetfilter-queue-dev libnetfilter-queue1
RUN apt-get install -y iptables ipset
RUN apt-get install -y sudo netcat iputils-ping wget
RUN apt-get autoremove -y && apt-get clean && rm -rf /var/lib/apt/lists/*

RUN wget -q https://dl.google.com/go/go${GOVERSION}.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go${GOVERSION}.linux-amd64.tar.gz && \
    rm go${GOVERSION}.linux-amd64.tar.gz

# Hacks for this issue https://github.com/AkihiroSuda/go-netfilter-queue/issues/5
RUN file='/usr/include/linux/netfilter.h' && \
    content="$(cat $file)" && \
    echo '#include <stdint.h>' >> "${file}" && \
    echo "${content}" >> $file


COPY ./ /app/

RUN mkdir -p /go/bin/; go build -o /go/bin/dockersnitch
RUN rm -rf /app

CMD ["/go/bin/dockersnitch"]
