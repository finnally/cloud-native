export tag=v1.0
export dockerhub_username=hellionc
export dockerhub_password=MTk4ODAxMzE
set env:
	export ROOT=github.com/cloud-native
	export GO111MODULE=off

build:
	echo "building httpserver binary"
	mkdir -p bin/amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/amd64 .

release: build
	echo "building httpserver container"
	docker build -t hellionc/httpserver:${tag} .

push: release
	echo "pushing image to hellionc/httpserver"
	docker login -u="${dockerhub_username}" -p="${dockerhub_password}"
	docker push hellionc/httpserver:${tag}
