IMAGE ?= wheatstalk/sftp-reporter
TAG ?= latest

all: build container

build:
	go build -o build/sftp-reporter main.go

container:
	docker build -t $(IMAGE):$(TAG) .