#!/usr/bin/env bash
set -ex
export GITHUB_TOKEN="a9e8fd7694ce5bfa6f009aa86150e949727571bc"
export DOCKER_USERNAME="kpianeta"
export DOCKER_PASSWORD="Diache123!"

docker pull goreleaser/goreleaser
function goreleasers { 
    docker run --rm --privileged \
        -v $(pwd):/go/src/github.com/user/repo \
        -v /var/run/docker.sock:/var/run/docker.sock \
        -w /go/src/github.com/user/repo \
        -e GITHUB_TOKEN=${GITHUB_TOKEN} \
        -e DOCKER_USERNAME=${DOCKER_USERNAME} \
        -e DOCKER_PASSWORD=${DOCKER_PASSWORD} \
        goreleaser/goreleaser "$@"
    }
goreleasers release --rm-dist --skip-publish