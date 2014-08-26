version=$(shell cat VERSION).$(shell git rev-list --count $(shell git log -1 --pretty=format:%H -- VERSION)..HEAD .)

.PHONY: version
version:
	@echo $(version)

ec2disks: ec2disks.go VERSION
	go build -ldflags "-X main.version $(version) -X main.buildDate $(shell date +'%Y-%m-%d_%H:%M:%S%z')" $<
