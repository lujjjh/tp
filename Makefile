all: build

build:
	GOOS=linux GOARCH=amd64 go build -o bin/tpredir tp/cmd/tpredir
	GOOS=linux GOARCH=amd64 go build -o bin/tpserver tp/cmd/tpserver
