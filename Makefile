TAG=`git describe --exact-match --tags $(git log -n1 --pretty='%h') 2>/dev/null || git rev-parse --abbrev-ref HEAD`
LDFLAGS=-ldflags "-X main.Version=${TAG}"
BIN_PATH=${GOPATH}/src/github.com/foomo/nodeflock/bin

all: test build
prepare: clean ts
	cp nodeflock-process/process.js bin/nodeflock-process.js
ts:
	tsc --out nodeflock-process/process.js nodeflock-process/process.ts
	chmod u+x nodeflock-process/process.js
run-ts-demo: ts
	nodeflock-process/process.js /Users/jan/go/src/github.com/foomo/nodeflock/nodeflock-process/demo.js
clean:
	rm -vf bin/*
build: prepare
	go build $(LDFLAGS) -o bin/nodeflock cmd/nodeflock/nodeflock.go
build-linux: prepare
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/nodeflock-linux cmd/nodeflock/nodeflock.go
run: build
	PATH=${PATH}:$(BIN_PATH) bin/nodeflock /Users/jan/www/globus/var/test/modules/MZG.Components/comps.js
