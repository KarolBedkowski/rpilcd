VERSION=`git describe --always`
DATE=`date`
LDFLAGS="-X k.prv/rpilcd.AppVersion '$(VERSION) - $(DATE)'"

.PHONY: resources build

build: resources
	GOGCCFLAGS="-s -fPIC -O4 -Ofast -march=native" godep go build -ldflags $(LDFLAGS)

build_pi: resources
	CGO_ENABLED="0" GOGCCFLAGS="-fPIC -O4 -Ofast -march=native -s" GOARCH=arm GOARM=5 go build -o rpilcd -ldflags $(LDFLAGS)
	#CGO_ENABLED="0" GOGCCFLAGS="-g -O2 -fPIC" GOARCH=arm GOARM=5 go build server.go 

clean:
	go clean
	rm -fr server rpilcd dist build
	find . -iname '*.orig' -delete

install_pi: build_pi
	scp rpilcd pi:

deps:
	go get -d -v ./...
