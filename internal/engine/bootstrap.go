package engine

import (
	"io/fs"
	"log"

	"gvn-engine/internal/loader"
	"gvn-engine/internal/script"
)

// Bootstrap initializes the game from an embedded assets FS.
// Shared by desktop main.go and mobile binding.
func Bootstrap(assetsFS fs.FS, scriptPath string) (*Game, error) {
	log.SetFlags(log.Ltime | log.Lshortfile)

	assets := loader.NewAssetManager(assetsFS)

	data, err := assets.ReadFile(scriptPath)
	if err != nil {
		return nil, err
	}

	sf, err := script.Parse(data)
	if err != nil {
		return nil, err
	}

	game := NewGame(sf, assets, assetsFS)
	// game.Ctx.ScriptFile = scriptPath // 暂时注释掉，因为 Context 结构体可能没更新 ScriptFile 字段？
	// 实际上 Context 结构体里有 ScriptFile。
	// 但是 NewContext 不需要参数。
	// 我们在这里手动设置。
	game.Ctx.ScriptFile = scriptPath
	return game, nil
}
