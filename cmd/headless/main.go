// cmd/headless: a standalone script validator with zero ebiten dependency.
// Safe to build and run in headless CI environments without X11 or GPU.
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"

	gvnengine "gvn-engine"
	"gvn-engine/internal/headless"
)

func main() {
	scriptPath := flag.String("script", "scripts/demo.json", "script file path (relative to assets/)")
	flag.Parse()

	assetsDir, err := fs.Sub(gvnengine.AssetsFS, "assets")
	if err != nil {
		log.Fatalf("assets FS error: %v", err)
	}

	data, err := fs.ReadFile(assetsDir, *scriptPath)
	if err != nil {
		log.Fatalf("cannot read %q: %v", *scriptPath, err)
	}

	res, err := headless.RunAuto(*scriptPath, data)
	if err != nil {
		log.Fatalf("execution error: %v", err)
	}

	fmt.Println(res)

	if len(res.Warnings) > 0 {
		fmt.Fprintf(os.Stderr, "[HEADLESS] %d warning(s)\n", len(res.Warnings))
		os.Exit(1)
	}
	fmt.Println("[HEADLESS] OK")
}
