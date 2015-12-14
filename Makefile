# Requirements before building:
#
# Get protoc (https://developers.google.com/protocol-buffers/)
#  $ sudo port install protobuf-c
#
# Get go dependencies:
#  $ make get-go-dependencies


# Running locall (See locserve: below for testing locally.)
#
# serve needs to have run install_html.
serve: generate 
	cd go; GOPATH=$$(pwd) go run src/toutizes.com/test/aserve.go --port=9090  --orig_root "~/Google Drive/Photos" --root=/tmp/db-full --static_root=/tmp/db/htdocs

install_html:
	bin/install-ttpic.sh

# Push to server
push_serve: generate
	cd go; GOPATH=$$(pwd) GOOS=linux GOARCH=amd64 go build -o bin_linux/aserve src/toutizes.com/test/aserve.go
	rsync "go/bin_linux/aserve" ec2:/mnt/photos/bin/

push_html:
	rsync --delete --recursive /tmp/db/htdocs/db/ ec2:/mnt/photos/htdocs/db/

# Drive sync.
drive:
	cd go; GOPATH=$$(pwd) GOBIN="bin" go install src/github.com/odeke-em/drive/cmd/drive/main.go; mv bin/main bin/drive

push_drive: 
	cd go; GOPATH=$$(pwd) GOOS=linux GOARCH=amd64 go build -o bin_linux/drive src/github.com/odeke-em/drive/cmd/drive/main.go
	rsync "go/bin_linux/drive" ec2:/mnt/photos/bin/

# Lightroom sync.
sync:
	rm -f /tmp/lrlog
	cd go; GOPATH=$$(pwd) go run src/toutizes.com/async.go /Users/matthieu/Pictures/Lightroom/Photos/2015/2015-01-18/final/*.jpg
	cat /tmp/lrlog

install_sync:
	cd go; GOPATH=$$(pwd) GOBIN="bin" go install src/toutizes.com/async.go
	cp go/bin/async "$(HOME)/Library/Application Support/Adobe/Lightroom/Export Actions/"

# Protocol buffers
STORE=src/toutizes.com/store

generate:
	mkdir -p go/$(STORE)
	(cd proto; PATH=$$PATH:../go/bin protoc --go_out=../go/$(STORE) db.proto)

clean:
	rm -f $(STORE)/*.pb.go

# Tests.
DIR=2005
run: generate
	cd go; GOPATH=$$(pwd) go run src/toutizes.com/test/run.go --orig_root ~/projects/test-photos/2002

locserve: generate
	cd go; GOPATH=$$(pwd) go run src/toutizes.com/test/aserve.go --orig_root ~/projects/test-photos --root /tmp/db-loc  --static_root=/tmp/db/htdocs --port 8080 --num_cpu 2

test: generate
	cd src/toutizes.com/model; go test


# Get dependencies for the Go code.
# Ignore error from github.com/rwcarlsen/goexif which is because it has no files to build.
get-go-dependencies:
	-cd go; GOPATH=$$(pwd) go get -d github.com/rwcarlsen/goexif
	cd go; GOPATH=$$(pwd) go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
