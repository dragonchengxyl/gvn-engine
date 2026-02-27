package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"

	gvnengine "gvn-engine"
	"gvn-engine/internal/engine"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	headless := flag.Bool("headless", false, "run in headless mode (no window, CI/test use)")
	scriptPath := flag.String("script", "scripts/chapter1.json", "script file path (relative to assets/)")
	flag.Parse()

	assetsDir, err := fs.Sub(gvnengine.AssetsFS, "assets")
	if err != nil {
		log.Fatalf("failed to open embedded assets: %v", err)
	}

	if *headless {
		runHeadless(assetsDir, *scriptPath)
		return
	}

	game, err := engine.Bootstrap(assetsDir, *scriptPath)
	if err != nil {
		log.Fatalf("failed to bootstrap: %v", err)
	}

	ebiten.SetWindowSize(960, 540)
	ebiten.SetWindowTitle("GVN-Engine")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatalf("engine error: %v", err)
	}
}

// runHeadless reads the script from the embedded FS and runs it without a window.
// Exits with code 1 if validation fails, 0 on success.
func runHeadless(assetsFS fs.FS, scriptPath string) {
	data, err := fs.ReadFile(assetsFS, scriptPath)
	if err != nil {
		log.Fatalf("[HEADLESS] cannot read %q: %v", scriptPath, err)
	}

	res, err := engine.RunHeadlessAuto(scriptPath, data)
	if err != nil {
		log.Fatalf("[HEADLESS] execution error: %v", err)
	}

	fmt.Println(res)

	if len(res.Warnings) > 0 {
		fmt.Fprintf(os.Stderr, "[HEADLESS] completed with %d warning(s)\n", len(res.Warnings))
		os.Exit(1)
	}
	fmt.Println("[HEADLESS] OK — script validated successfully")
}
