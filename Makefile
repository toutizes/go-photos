SERVER=ce1
LR_ACTIONS="$(HOME)/Library/Application Support/Adobe/Lightroom/Export Actions"
SYN_TXT=backend/model/synonyms.txt

# Toutizes.
# serve_small needs to have run install_html.
# http://localhost:8080
serve_small: generate install_flutter
	cp ~/Google\ Drive/My\ Drive/Floutizes/floutizes-firebase-adminsdk-fbsvc-584c308c82.json /tmp/aserve/db/
	go run backend/test/aserve.go --bin_root=/opt/homebrew/bin/ --use_https=false --orig_root="/Users/matthieu/projects/test-photos" --root=/tmp/aserve/db-full --static_root=/tmp/aserve/db/htdocs --firebase_creds=/tmp/aserve/db/floutizes-firebase-adminsdk-fbsvc-584c308c82.json

install_html:
	cmd/install-ttpic.sh

install_flutter: 
	rsync --delete --recursive floutizes/build/web/* /tmp/aserve/db/htdocs/flutter
	rsync --delete --recursive $(SYN_TXT) $(SERVER):/mnt/photos/htdocs/synonyms.txt

push_serve: generate 
	GOOS=linux GOARCH=amd64 go build -o bin_linux/aserve backend/test/aserve.go
	rsync "bin_linux/aserve" $(SERVER):/mnt/photos/bin/

push_html: install_html
	rsync --delete --recursive /tmp/aserve/db/htdocs/db/ $(SERVER):/mnt/photos/htdocs/db/
	rsync --delete --recursive $(SYN_TXT) $(SERVER):/mnt/photos/htdocs/synonyms.txt

push_flutter: 
	rsync --delete --recursive floutizes/build/web/* $(SERVER):/mnt/photos/htdocs/flutter/
	rsync --delete --recursive $(SYN_TXT) $(SERVER):/mnt/photos/htdocs/synonyms.txt
	rsync ~/Google\ Drive/My\ Drive/Floutizes/floutizes-firebase-adminsdk-fbsvc-584c308c82.json $(SERVER):/mnt/photos/

run: generate
	go run backend/test/run.go --orig_root /Volumes/GoogleDrive/Mon\ Drive/Photos/1997

minify: 
	go run backend/util/minify.go --bin_root=/opt/homebrew/bin/ --orig_root="/Users/matthieu/projects/test-photos" --root=/tmp/aserve/db-full

push_minify:
	GOOS=linux GOARCH=amd64 go build -o bin_linux/minify backend/util/minify.go
	rsync "bin_linux/minify" $(SERVER):/mnt/photos/bin/

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
install_sync:
	go build --ldflags='-X main.Type=F' -o $(LR_ACTIONS)/gsync sync/gsync.go
	go build --ldflags='-X main.Type=I' -o $(LR_ACTIONS)/gsync_incr sync/gsync.go

test_sync:
	go run --ldflags='-X main.Type=I' sync/gsync.go '/Users/matthieu/Pics from Bernard/Bernard/1953 and before/final/fianchabi.jpg'

install_sync_inc:
	go build -o bin/gsync sync/gsync.go
	cp bin/gsync "$(HOME)/Library/Application Support/Adobe/Lightroom/Export Actions/gsync_inc"

# Rules
%.pb.go: %.proto
	protoc --go_out=. $<
