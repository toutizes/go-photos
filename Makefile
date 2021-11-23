SERVER=ce1

# Toutizes.
# serve_small needs to have run install_html.
serve_small: generate install_html
	go run backend/test/aserve.go --bin_root=/opt/homebrew/bin/ --use_https=false --orig_root="/Users/matthieu/projects/test-photos" --root=/tmp/aserve/db-full --static_root=/tmp/aserve/db/htdocs

install_html:
	cmd/install-ttpic.sh

push_serve: generate push_html
	GOOS=linux GOARCH=amd64 go build -o bin_linux/aserve backend/test/aserve.go
	rsync "bin_linux/aserve" $(SERVER):/mnt/photos/bin/

push_html:
	rsync --delete --recursive /tmp/db/htdocs/db/ $(SERVER):/mnt/photos/htdocs/db/

run: generate
	go run backend/test/run.go --orig_root /Volumes/GoogleDrive/Mon\ Drive/Photos/1997

addkwds: generate
	go run backend/test/addkwds.go --root /tmp/a

GENERATED=backend/store/db.pb.go
generate: $(GENERATED)

clean:
	go clean
	rm -f $(GENERATED)

# Google drive.
push_drive:
	GOBIN=$$(pwd)/bin go get -u github.com/odeke-em/drive/cmd/drive
	GOOS=linux GOARCH=amd64 go build -o bin_linux/drive github.com/odeke-em/drive/cmd/drive
	rsync "bin_linux/drive" $(SERVER):/mnt/photos/bin/
	rsync "cmd/fix-dir-time.py" $(SERVER):/mnt/photos/bin/

# Lightroom sync.
sync:
	rm -f /tmp/lrlog
	go run sync/async.go /Users/matthieu/Pictures/Lightroom/Photos/2015/2015-01-18/final/*.jpg
	cat /tmp/lrlog

install_sync:
	go build -o bin/gsync sync/gsync.go
	cp bin/gsync "$(HOME)/Library/Application Support/Adobe/Lightroom/Export Actions/gsync"

install_sync_inc:
	go build -o bin/gsync sync/gsync.go
	cp bin/gsync "$(HOME)/Library/Application Support/Adobe/Lightroom/Export Actions/gsync_inc"

# Rules
%.pb.go: %.proto
	protoc --go_out=. $<
