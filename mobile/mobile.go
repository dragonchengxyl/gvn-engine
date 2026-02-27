package mobile

import (
	"io/fs"
	"log"

	gvnengine "gvn-engine"
	"gvn-engine/internal/engine"

	"github.com/hajimehoshi/ebiten/v2/mobile"
)

func init() {
	assetsDir, err := fs.Sub(gvnengine.AssetsFS, "assets")
	if err != nil {
		log.Fatalf("mobile: failed to open embedded assets: %v", err)
	}

	game, err := engine.Bootstrap(assetsDir, "scripts/demo.json")
	if err != nil {
		log.Fatalf("mobile: failed to bootstrap: %v", err)
	}

	mobile.SetGame(game)
}

// Dummy is a placeholder to avoid "no exported functions" error from gomobile.
func Dummy() {}
