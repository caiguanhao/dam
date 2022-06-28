DDD := $(shell date +"%Y%m%d")
VERSION := $(shell git rev-parse --short HEAD)

build:
	GOOS=linux GOARCH=arm GOARM=7 go build -v -ldflags="-X 'main.version=$(VERSION)'" -o dam
	tar cfvz dam-$(DDD)-$(VERSION)-arm.tar.gz dam
	GOOS=linux GOARCH=amd64 go build -v -ldflags="-X 'main.version=$(VERSION)'" -o dam
	tar cfvz dam-$(DDD)-$(VERSION)-amd64.tar.gz dam

hash:
	@echo amd64 md5
	@tar -Oxzf dam-$(DDD)-$(VERSION)-amd64.tar.gz dam | openssl md5
	@echo arm md5
	@tar -Oxzf dam-$(DDD)-$(VERSION)-arm.tar.gz dam | openssl md5

clean:
	rm -f dam*.tar.gz
