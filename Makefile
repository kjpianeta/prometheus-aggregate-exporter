MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
MKFILE_DIR := $(dir $(MKFILE_PATH))
DISTDIR := $(MKFILE_DIR)dist/
RPMS := $(shell find $(DISTDIR) -name 'prometheus-aggregate-exporter-*el6.x86_64.rpm')

rpm:
	@echo "building: $@"
	goreleaser/goreleaser release --rm-dist \

test:
	$(info Make: Testing RPM.)
	echo ${RPMS} && cd rpm/test &&  ./run.sh ${RPMS}

clean:
	$(info Make: Cleaning RPM.)

.PHONY: rpm test clean test tag release install uninstall all