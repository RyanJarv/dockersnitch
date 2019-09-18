#!/usr/bin/env bash
docker run -it -v $(pwd):/go/src/github.com/RyanJarv/dockersnitch -p 127.0.0.1:33504:33504 -w /go/src/github.com/RyanJarv/dockersnitch --privileged --pid host dockersnitch sh -c "go run main.go & bash -l"
