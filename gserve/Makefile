GOPATH := ${PWD}:${GOPATH}
export GOPATH

default: build

build:
	go get -d -v ./src/gserve
	go build -tags netgo -v -o ./bin/gserve ./src/gserve
