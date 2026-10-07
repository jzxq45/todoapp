.PHONY: help run open dev test build lint docker docker-build docker-run docker-stop docker-logs deps-android mobile-bind apk clean

GOMOBILE_VERSION = v0.0.0-20241016134751-7ff83004ec2c
GOMOBILE_PACKAGE = golang.org/x/mobile/cmd/gomobile@$(GOMOBILE_VERSION)

ifeq ($(OS),Windows_NT)
GRADLEW = gradlew.bat
GRADLE_COMMAND = $(GRADLEW)
FORMAT_CHECK_COMMAND = powershell -NoProfile -Command "$$files = gofmt -l .; if ($$files) { gofmt -d .; exit 1 }"
CLEAN_COMMAND = powershell -NoProfile -Command "Remove-Item -LiteralPath 'bin','coverage.out','todo.db','android\.gradle','android\app\build','android\app\libs\todo.aar','android\app\libs\todo-sources.jar' -Force -Recurse -ErrorAction SilentlyContinue"
APK_PREP_COMMAND = powershell -NoProfile -Command "New-Item -ItemType Directory -Force -Path 'dist' | Out-Null; Remove-Item -LiteralPath 'dist\TodoApp.apk' -Force -ErrorAction SilentlyContinue"
APK_COPY_COMMAND = powershell -NoProfile -Command "Copy-Item -LiteralPath 'android\app\build\outputs\apk\debug\app-debug.apk' -Destination 'dist\TodoApp.apk' -Force"
else
GRADLEW = gradlew
GRADLE_COMMAND = sh ./$(GRADLEW)
FORMAT_CHECK_COMMAND = if [ -n "$$(gofmt -l .)" ]; then gofmt -d .; exit 1; fi
CLEAN_COMMAND = rm -rf bin coverage.out todo.db android/.gradle android/app/build android/app/libs/todo.aar android/app/libs/todo-sources.jar
APK_PREP_COMMAND = mkdir -p dist && rm -f dist/TodoApp.apk
APK_COPY_COMMAND = cp android/app/build/outputs/apk/debug/app-debug.apk dist/TodoApp.apk
endif

help:
	@echo "TodoApp commands:"
	@echo "  make run          Build and run the web app at http://localhost:8080"
	@echo "  make open         Start the app and open the browser"
	@echo "  make dev          Run with automatic Go source reload"
	@echo "  make test         Run all tests"
	@echo "  make lint         Run gofmt and go vet"
	@echo "  make docker       Build and run with Docker Compose"
	@echo "  make deps-android Install gomobile tooling (Android SDK/JDK/NDK required)"
	@echo "  make mobile-bind  Build the Android AAR from the Go backend"
	@echo "  make apk          Build a fresh APK at dist/TodoApp.apk"

run:
	go run ./cmd/server

open:
	go run ./cmd/open

dev:
	go run ./cmd/dev

test:
	go test -race ./...

build:
	go build -trimpath -o bin/todoapp ./cmd/server

lint:
	$(FORMAT_CHECK_COMMAND)
	go vet ./...

docker:
	docker compose up --build

docker-build:
	docker compose build

docker-run:
	docker compose up -d --build

docker-stop:
	docker compose down

docker-logs:
	docker compose logs -f app

deps-android:
	go install $(GOMOBILE_PACKAGE)
	go run $(GOMOBILE_PACKAGE) init

mobile-bind: deps-android
	go run $(GOMOBILE_PACKAGE) bind -target=android -androidapi 24 -javapkg=com.example.todoapp -o android/app/libs/todo.aar ./mobile

apk:
	$(APK_PREP_COMMAND)
	$(MAKE) mobile-bind
	cd android && $(GRADLE_COMMAND) --no-daemon clean assembleDebug
	$(APK_COPY_COMMAND)

clean:
	go clean ./...
	$(CLEAN_COMMAND)
