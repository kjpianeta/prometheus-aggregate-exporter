ifndef GITHUB_TOKEN
$(error The GITHUB_TOKEN environment variable is missing.)
endif

rpm:
	@echo "building: $@"
	@docker run \
	    -ti \
	    --rm \
	    --privileged \
	    -v "$$(pwd):/go/src/github.com/user/repo" \
	    -v /var/run/docker.sock:/var/run/docker.sock \
	    -w /go/src/github.com/user/repo \
	    -e "GITHUB_TOKEN=$(GITHUB_TOKEN)" \
	    goreleaser/goreleaser release --rm-dist \

test:
	$(info Make: Testing RPM.)
	cd rpm/test && ./run.sh

.PHONY: rpm test clean test tag release install uninstall all