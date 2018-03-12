#!/usr/bin/env bash
docker run -it -v $(pwd):/go/src/github.com/RyanJarv/dockersnitch -w /go/src/github.com/RyanJarv/dockersnitch --privileged --pid host dockersnitch bash -l
