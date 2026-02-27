# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

GVN-Engine — a data-driven, cross-platform (Windows + Android) visual novel engine built with Go and Ebitengine v2.

## Build & Run

```bash
go run ./cmd/game           # run desktop
go build -o gvn-engine.exe ./cmd/game  # build desktop
go build ./...              # check all packages compile
go vet ./...                # static analysis
go test ./...               # run all tests
make android                # build Android APK (requires ebitenmobile + Android SDK/NDK)
```

## Architecture

- `embed.go` (root) — embeds `assets/` via `//go:embed`; imported as `gvnengine "gvn-engine"`
- `cmd/game/main.go` — desktop entry point
- `mobile/mobile.go` — Android entry point via `ebitenmobile bind`
- `internal/engine/` — core game loop, state machine, runtime context, bootstrap
- `internal/script/` — JSON script parser with skip-on-error, label indexing for jumps
- `internal/loader/` — `embed.FS` wrapper with image caching + placeholder fallback + `GenerateRect`
- `internal/render/` — multi-layer renderer (BG/chars/FG), text renderer with CJK auto-wrap, choice UI
- `internal/input/` — unified mouse/touch with just-pressed detection
- `internal/audio/` — BGM (looped) + SE (one-shot), OGG/WAV/MP3, silent fallback
- `assets/` — game resources embedded at compile time

## Key Constraints (from SKILL.md)

- No `os.Open` or absolute paths — all resource access through `embed.FS` / `io/fs`
- No `panic()` — return errors; placeholder images for missing resources; silent audio fallback
- `Draw()` must be stateless — no mutation, pure function of game state
- No heavy I/O in `Update()`/`Draw()` — preload or use goroutines
- Virtual resolution 1920x1080, scaled via `GeoM`
- Script errors are logged and skipped, never crash the engine
- All code must be "out-of-the-box" runnable — use `GenerateRect` / built-in gofont as fallbacks
