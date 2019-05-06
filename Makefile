.PHONY: build clean deploy

build:
	dep ensure -v
	env GOOS=linux go build -ldflags="-s -w" -o bin/parser parser/main.go

clean:
	rm -rf ./bin ./vendor Gopkg.lock

deploy: build
	sls deploy --verbose
