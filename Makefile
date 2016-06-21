TAG=`git describe --exact-match --tags $(git log -n1 --pretty='%h') 2>/dev/null || git rev-parse --abbrev-ref HEAD`
#LDFLAGS='-ldflags -X main.Version=$(TAG)'
LDFLAGS=-ldflags "-X main.Version=${TAG}"

all: test build
ts:
	tsc --out nodeflock-process/process.js nodeflock-process/process.ts
	chmod u+x nodeflock-process/process.js
run-ts-demo: ts
	nodeflock-process/process.js /Users/jan/go/src/github.com/foomo/nodeflock/nodeflock-process/demo.js
clean:
	echo "nothing to clean yet"
build: prepare
	go build $(LDFLAGS) config-bob.go
build-arch: prepare
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/config-bob-linux-amd64_$(TAG) config-bob.go
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/config-bob-darwin-amd64_$(TAG) config-bob.go