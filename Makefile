VERSION:=$(shell git describe --always)
DATE:=$(shell date +%Y%m%d%H%M%S)
VERS:=$(VERSION)-$(DATE)
LDFLAGS:="-X main.AppVersion='$(VERS)' -w -s"
# -gcflags "-N -l"

.PHONY: build

build: 
	GOGCCFLAGS="-s -fPIC -O4 -Ofast -march=native" go build -ldflags "$(LDFLAGS)" -v
	#GOGCCFLAGS="-g -s -march=native" go build -ldflags $(LDFLAGS)

build_pi: rpilcd_pi

rpilcd_pi: *.go
	GOGCCFLAGS="-fPIC -O4 -Ofast -march=native -pipe -mcpu=arm1176jzf-s -mfpu=vfp -mfloat-abi=hard -s" \
		CHOST="armv6j-hardfloat-linux-gnueabi" \
		CXX=arm-linux-gnueabihf-g++ CC=arm-linux-gnueabihf-gcc \
		GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=1 \
		go build -v --ldflags '-extldflags "-static"' --ldflags $(LDFLAGS) -o rpilcd_pi

#	CGO_ENABLED="0" GOGCCFLAGS="-fPIC -O4 -Ofast -march=native -s" GOARCH=arm GOARM=6 go build -o rpilcd_pi -ldflags "$(LDFLAGS) -w" -v
	#CGO_ENABLED="0" GOGCCFLAGS="-g -fPIC -march=native -s" GOARCH=arm GOARM=5 go build -o rpilcd -ldflags $(LDFLAGS)

run: 
	go run *.go -console=true -alsologtostderr

clean:
	go clean
	rm -fr server rpilcd dist build
	find . -iname '*.orig' -delete

install_pi: rpilcd_pi
	ssh pi sudo service rpilcd stop
	scp rpilcd_pi pi:rpilcd/rpilcd
	ssh pi sudo service rpilcd start

deps:
	go get -d -v ./...
