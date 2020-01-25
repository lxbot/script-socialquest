.PHONY: build

build:
	go build -buildmode=plugin -o script-socialquest.so script.go
