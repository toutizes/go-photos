SERVER=ce1
LR_ACTIONS="$(HOME)/Library/Application Support/Adobe/Lightroom/Export Actions"
SYN_TXT=backend/catalog/synonyms.txt

# Toutizes.
# serve_small needs to have run install_html.
# http://localhost:8080
serve_small: generate install-web
	cp ~/Google\ Drive/My\ Drive/Installations/Floutizes/floutizes-firebase-adminsdk-fbsvc-584c308c82.json /tmp/aserve/db/
	mkdir -p /tmp/aserve/log
	go run ./backend/cmd/aserve --bin_root=/opt/homebrew/bin/ --use_https=false --orig_root="/Users/matthieu/projects/test-photos" --root=/tmp/aserve/db-full --static_root=/tmp/aserve/db/htdocs --firebase_creds=/tmp/aserve/db/floutizes-firebase-adminsdk-fbsvc-584c308c82.json --log_dir=/tmp/aserve/log --use_lr_parser # &> /tmp/aserve/log/aserve.log

install_html:
	cmd/install-ttpic.sh

install_flutter: 
	mkdir -p /tmp/aserve/db/htdocs/flutter
	rsync --delete --recursive floutizes/build/web/* /tmp/aserve/db/htdocs/flutter
	rsync --delete --recursive $(SYN_TXT) /tmp/aserve/db/htdocs/synonyms.txt

push_serve: generate 
	GOOS=linux GOARCH=amd64 go build -o bin_linux/aserve ./backend/cmd/aserve
	rsync "bin_linux/aserve" $(SERVER):/mnt/photos/bin/

push_html: install_html
	rsync --delete --recursive /tmp/aserve/db/htdocs/db/ $(SERVER):/mnt/photos/htdocs/db/
	rsync --delete --recursive $(SYN_TXT) $(SERVER):/mnt/photos/htdocs/synonyms.txt


push-flutter-release:
	(cd floutizes; make build-web-release)
	rsync --delete --recursive floutizes/build/web/* $(SERVER):/mnt/photos/htdocs/flutter/
	rsync --delete --recursive $(SYN_TXT) $(SERVER):/mnt/photos/htdocs/synonyms.txt
	rsync ~/Google\ Drive/My\ Drive/Floutizes/floutizes-firebase-adminsdk-fbsvc-584c308c82.json $(SERVER):/mnt/photos/

# React frontend: same Go static directory, normal Firebase authentication.
.PHONY: build-web-release install-web push-web-release push-flutter-release
build-web-release:
	pnpm --dir frontend build

install-web: build-web-release
	mkdir -p /tmp/aserve/db/htdocs/flutter
	rsync --recursive frontend/dist/ /tmp/aserve/db/htdocs/flutter/
	rsync --delete --recursive $(SYN_TXT) /tmp/aserve/db/htdocs/synonyms.txt

# Keep previous hashed assets for already-open tabs; update the entry point last.
push-web-release: build-web-release
	rsync --recursive --exclude=index.html frontend/dist/ $(SERVER):/mnt/photos/htdocs/flutter/
	rsync frontend/dist/index.html $(SERVER):/mnt/photos/htdocs/flutter/index.html
	rsync $(SYN_TXT) $(SERVER):/mnt/photos/htdocs/synonyms.txt

# Optional React trial alongside production, using the existing Go /app/ route.
# Use a named HTML entry for correct MIME; print a fresh query to bypass its cache.
.PHONY: build-web-test push-web-test release-react-app
build-web-test:
	pnpm --dir frontend exec tsc -b
	pnpm --dir frontend exec vite build --mode staging --outDir dist-test

release-react-app: push-web-release

push-web-test: build-web-test
	ssh $(SERVER) 'mkdir -p /mnt/photos/htdocs/flutter/react-test'
	rsync --recursive --exclude=index.html --exclude=flutter_service_worker.js frontend/dist-test/ $(SERVER):/mnt/photos/htdocs/flutter/react-test/
	rsync frontend/dist-test/index.html $(SERVER):/mnt/photos/htdocs/flutter/react-test.html
	@echo "React test URL: https://toutizes.com/app/react-test.html?v=$$(date +%s)"

run: generate
	go run backend/test/run.go --orig_root /Volumes/GoogleDrive/Mon\ Drive/Photos/1997

minify: 
	go run ./backend/cmd/minify --bin_root=/opt/homebrew/bin/ --orig_root="/Users/matthieu/projects/test-photos" --root=/tmp/aserve/db-full

push_minify:
	GOOS=linux GOARCH=amd64 go build -o bin_linux/minify ./backend/cmd/minify
	rsync "bin_linux/minify" $(SERVER):/mnt/photos/bin/

list_kwds: generate
	go run ./backend/cmd/list_kwds /Users/matthieu/projects/test-photos/2003/2003-05-07/Image02.jpg

# go run backend/test/list_kwds.go ~/projects/test-photos/2003/2003-05-07/Image06.jpg 
# go run backend/test/list_kwds.go ~/projects/test-photos/1979/Devin-io-0168.jpg 

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

build_sync:
	go build --ldflags='-X main.Type=F' -o bin/gsync sync/gsync.go
	go build --ldflags='-X main.Type=I' -o bin/gsync_incr sync/gsync.go

test_sync:
	go run --ldflags='-X main.Type=I' sync/gsync.go --dry_run /Users/matthieu/Pictures/Photos/2025/2025-06-11/final/PXL_20250611_233024990.jpg /Users/matthieu/Pictures/Photos/2025/2025-06-11/final/PXL_20250612_002002574.jpg

# Rules
%.pb.go: %.proto
	protoc --go_out=. $<
