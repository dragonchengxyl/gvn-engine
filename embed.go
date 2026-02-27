package gvnengine

import "embed"

// AssetsFS embeds the entire assets directory at compile time.
// This file must live at the module root so the embed path resolves correctly.
//
//go:embed assets
var AssetsFS embed.FS
