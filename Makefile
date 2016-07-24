PHONY: prepare build run

prepare:
	docker build -t mc-dev - < Dockerfile.dev

build:
