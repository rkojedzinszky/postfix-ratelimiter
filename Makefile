all: postfix-ratelimiter

VERSION = $(shell git describe --tags)

GO = go

postfix-ratelimiter:
	$(GO) build -ldflags "-s -X main.version=$(VERSION)" .

clean:
	rm -f postfix-ratelimiter
