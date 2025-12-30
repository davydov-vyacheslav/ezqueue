.PHONY: run build test docker-build docker-run clean

run:
	go run .

clean:
	rm -rf ./output

prepare:
	mkdir output

build: clean prepare
	swag init
	go build -o output/server .

test:
	go test -v ./...

docker-build:
	docker build -t ezqueue .