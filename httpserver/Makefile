export tag=v2.0
export dockerhub_username=hellionc
export dockerhub_password=MTk4ODAxMzE

build: 
	printf '\033[1;32;40m%b\033[0m\n' "building httpserver container"
	docker build -t hellionc/httpserver:${tag} .

push: build
	printf '\033[1;32;40m%b\033[0m\n' "pushing image to hellionc/httpserver"
	docker login -u="${dockerhub_username}" -p="${dockerhub_password}"
	docker push hellionc/httpserver:${tag}
