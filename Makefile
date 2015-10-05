VERSION=`git describe --always`
DATE=`date`
LDFLAGS=-X main.AppVersion '$(VERSION) - $(DATE)'
# -gcflags "-N -l"

.PHONY: build

build: 
	GOGCCFLAGS="-s -fPIC -O4 -Ofast -march=native" go build -ldflags "$(LDFLAGS)"
	#GOGCCFLAGS="-g -s -march=native" go build -ldflags $(LDFLAGS)

build_pi: 
	CGO_ENABLED="0" GOGCCFLAGS="-fPIC -O4 -Ofast -march=native -s" GOARCH=arm GOARM=5 go build -o rpilcd -ldflags "$(LDFLAGS) -w"
	#CGO_ENABLED="0" GOGCCFLAGS="-g -fPIC -march=native -s" GOARCH=arm GOARM=5 go build -o rpilcd -ldflags $(LDFLAGS)

run: 
	go run *.go -console=true

clean:
	go clean
	rm -fr server rpilcd dist build
	find . -iname '*.orig' -delete

install_pi: build_pi
	scp rpilcd pi:

deps:
	go get -d -v ./...
