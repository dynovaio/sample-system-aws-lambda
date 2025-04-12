export GOARCH=amd64
export GOOS=linux
export CGO_ENABLED=0

.PHONY: build-HelloWorldFunction

build-HelloWorldFunction:
	go build -ldflags="-s -w" -o bootstrap main.go
	cp bootstrap $(ARTIFACTS_DIR)
	cp collector.yaml $(ARTIFACTS_DIR)
	rm -rf bin
