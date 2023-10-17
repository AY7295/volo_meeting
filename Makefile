.PHONY: all clean buildx up

IMAGE_NAME=ay7295/volo
IMAGE_TAG=latest

all: up

buildx:
	docker buildx build --platform linux/amd64,linux/arm64 -t $(IMAGE_NAME):$(IMAGE_TAG) --push .

up:
	docker compose up -d

clean:
	docker compose down


