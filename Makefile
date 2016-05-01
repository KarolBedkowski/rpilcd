VERSION:=$(shell git describe --always)
DATE:=$(shell date)
VERS:=$(VERSION) - $(DATE)
LDFLAGS:=-X main.AppVersion '$(VERS)' -w -s
# -gcflags "-N -l"

.PHONY: build

build: 
	GOGCCFLAGS="-s -fPIC -O4 -Ofast -march=native" go build -ldflags "$(LDFLAGS)" -v
	#GOGCCFLAGS="-g -s -march=native" go build -ldflags $(LDFLAGS)

build_pi: 
	CGO_ENABLED="0" GOGCCFLAGS="-fPIC -O4 -Ofast -march=native -s" GOARCH=arm go build -o rpilcd -ldflags "$(LDFLAGS) -w" -v
	#CGO_ENABLED="0" GOGCCFLAGS="-g -fPIC -march=native -s" GOARCH=arm GOARM=5 go build -o rpilcd -ldflags $(LDFLAGS)

run: 
	go run *.go -console=true -alsologtostderr

clean:
	go clean
	rm -fr server rpilcd dist build
	find . -iname '*.orig' -delete

install_pi: build_pi
	ssh pi sudo service k_rpilcd stop
	scp rpilcd pi:
	ssh pi sudo service k_rpilcd start

deps:
	go get -d -v ./...
