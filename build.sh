#!/usr/bin/env bash
: "${GITHUB_TOKEN?Not set. Need to set env variable GITHUB_TOKEN}"
: "${DOCKER_USERNAME?Not set. Need to set env variable DOCKER_USERNAME}"
: "${DOCKER_PASSWORD?Not set. Need to set env variable DOCKER_PASSWORD}"

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
goreleasers release --rm-dist