package render

import (
	"gvn-engine/internal/ecs"

	"github.com/hajimehoshi/ebiten/v2"
)

// DrawECSLayer iterates the ECS world and draws all entities that have both
// a Sprite and an Alpha component. If the entity also has a ShaderFX component
// its compiled shader is applied via DrawRectShader; otherwise plain DrawImage.
//
// This function is called from Game.Draw as an additional rendering pass after
// the standard Renderer.Draw, enabling ECS-driven animated effects without
// replacing the existing rendering pipeline.
func DrawECSLayer(screen *ebiten.Image, world *ecs.World) {
	world.Each(func(e *ecs.Entity) {
		if e.Sprite == nil || e.Alpha == nil {
			return
		}
		img, ok := e.Sprite.Image.(*ebiten.Image)
		if !ok || img == nil {
			return
		}
		alpha := e.Alpha.Value
		if alpha <= 0 {
			return
		}

		// Determine draw position / scale from Transform (or use full-screen defaults)
		var geoM ebiten.GeoM
		if e.Transform != nil {
			geoM.Scale(e.Transform.ScaleX, e.Transform.ScaleY)
			geoM.Translate(e.Transform.X, e.Transform.Y)
		} else {
			// Scale to fill virtual resolution (same logic as Renderer.drawImageScaled)
			bw := float64(img.Bounds().Dx())
			bh := float64(img.Bounds().Dy())
			geoM.Scale(float64(VirtualWidth)/bw, float64(VirtualHeight)/bh)
		}

		// Shader path
		if e.ShaderFX != nil {
			shader, ok := e.ShaderFX.ShaderSrc.(*ebiten.Shader)
			if ok && shader != nil {
				drawWithShader(screen, img, shader, e.ShaderFX.Uniforms, geoM, alpha)
				return
			}
		}

		// Plain path
		op := &ebiten.DrawImageOptions{}
		op.GeoM = geoM
		op.ColorScale.ScaleAlpha(float32(alpha))
		screen.DrawImage(img, op)
	})
}

// drawWithShader draws src onto dst using the given compiled Kage shader.
func drawWithShader(
	dst, src *ebiten.Image,
	shader *ebiten.Shader,
	uniforms map[string]any,
	geoM ebiten.GeoM,
	alpha float64,
) {
	w, h := src.Bounds().Dx(), src.Bounds().Dy()

	// Build the destination rectangle after GeoM transform
	// (DrawRectShader draws into a rect on dst; we approximate with a scaled blit)
	op := &ebiten.DrawRectShaderOptions{}
	op.GeoM = geoM
	op.ColorScale.ScaleAlpha(float32(alpha))
	op.Images[0] = src
	if uniforms != nil {
		op.Uniforms = uniforms
	}
	dst.DrawRectShader(w, h, shader, op)
}
