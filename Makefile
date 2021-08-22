APP_NAME=pod-annotator
IMAGE_NAME=gvitzi/pod-annotator

build:
	go build -o pod-annotator

docker-build:
	docker build . --tag ${IMAGE_NAME}

docker-push:
	docker push ${IMAGE_NAME}

run:
	./pod-annotator --pod-name pod-annotator-example -namespace default -dir-path .annotations-test -v 5