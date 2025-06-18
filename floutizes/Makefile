build: FORCE
	flutter build web

#  --base-href /app/ not needed as /app/ is hardcoded.
build-web-release:
	flutter build web --release --dart-define BACKEND=https://toutizes.com

run-web:
	flutter run -d chrome --web-port 3000

run-web-tt:
	flutter run -d chrome --web-port 3000 --dart-define BACKEND=https://toutizes.com

run-web-tt-release:
	flutter run --release -d chrome --web-port 3000 --dart-define BACKEND=https://toutizes.com

run-ios-simulator-pro15:
	flutter run -d "iPhone 15 Pro Max"

run-ios-simulator-pro15-tt:
	flutter run -d "iPhone 15 Pro Max" --dart-define BACKEND=https://toutizes.com

run-pixel8:
	flutter run -d 3A101FDJH0099F --dart-define BACKEND=https://toutizes.com

run-pixel8-release:
	flutter run -d 3A101FDJH0099F --dart-define BACKEND=https://toutizes.com --release


launch-emulator:
	flutter emulators --launch Apple

FORCE:
