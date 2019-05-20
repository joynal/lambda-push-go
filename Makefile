.PHONY: build clean deploy

build:
	dep ensure -v
	env GOOS=linux go build -ldflags="-s -w" -o bin/parser parser/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/sender sender/main.go

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
