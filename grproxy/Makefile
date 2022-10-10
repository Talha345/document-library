GOPATH := ${PWD}:${GOPATH}
export GOPATH

default: build

build:
	go get -d -v ./src/grproxy
	go build -tags netgo -v -o ./bin/grproxy ./src/grproxy
