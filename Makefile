IMAGE ?= wheatstalk/sftp-reporter
TAG ?= latest

all: build container

build: main.go
	go build -o build/sftp-reporter main.go

container:
	docker build -t $(IMAGE):$(TAG) .