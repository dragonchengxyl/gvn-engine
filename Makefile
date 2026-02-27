.PHONY: run build android clean

# Desktop (Windows/Linux/macOS)
run:
	go run ./cmd/game

build:
	go build -o gvn-engine.exe ./cmd/game

# Android APK via ebitenmobile
# Prerequisites:
#   go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@latest
#   Android SDK + NDK installed, ANDROID_HOME set
ANDROID_OUT ?= gvn-engine.apk

android:
	ebitenmobile bind -target android -javapkg com.gvnengine.game -o $(ANDROID_OUT) ./mobile

# Cleanup
clean:
	rm -f gvn-engine.exe $(ANDROID_OUT)
	go clean -cache

# Vet & test
check:
	go vet ./...
	go build ./...

test:
	go test ./...
