.PHONY: run build test docker-build docker-run clean

run:
	go run .

clean:
	rm -f output

prepare:
	mkdir output

build: clean prepare
	go build -o output/server .

test:
	go test -v ./...

docker-build:
	docker build -t ezqueue .