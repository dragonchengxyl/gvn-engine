package ecs

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

// ── Built-in Kage shader sources ────────────────────────────────────────────

// ShaderGrayscale desaturates a sprite (used for "flashback" or "memory" scenes).
const ShaderGrayscale = `//kage:unit pixels
package main

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	clr := imageSrc0At(srcPos)
	gray := 0.299*clr.r + 0.587*clr.g + 0.114*clr.b
	return vec4(gray, gray, gray, clr.a) * color
}
`

// ShaderSepia applies a warm sepia tone (used for "past" scenes).
const ShaderSepia = `//kage:unit pixels
package main

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	clr := imageSrc0At(srcPos)
	gray := 0.299*clr.r + 0.587*clr.g + 0.114*clr.b
	r := clamp(gray*1.2+0.11, 0.0, 1.0)
	g := clamp(gray*1.0+0.0,  0.0, 1.0)
	b := clamp(gray*0.8-0.08, 0.0, 1.0)
	return vec4(r, g, b, clr.a) * color
}
`

// ShaderFadeBlack fades between the sprite and pure black.
// Uniform: Progress float (0.0=original, 1.0=black).
const ShaderFadeBlack = `//kage:unit pixels
package main

var Progress float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	clr := imageSrc0At(srcPos)
	return mix(clr, vec4(0, 0, 0, clr.a), Progress) * color
}
`

// ── Shader registry ─────────────────────────────────────────────────────────

// ShaderName identifies a built-in shader.
type ShaderName string

const (
	ShaderNameGrayscale  ShaderName = "grayscale"
	ShaderNameSepia      ShaderName = "sepia"
	ShaderNameFadeBlack  ShaderName = "fade_black"
)

// shaderCache stores compiled *ebiten.Shader instances keyed by ShaderName.
// Shaders are compiled once and reused.
var shaderCache = map[ShaderName]*ebiten.Shader{}

// GetBuiltinShader returns (or lazily compiles) a named built-in Kage shader.
// Returns (nil, error) if compilation fails — callers should fall back to plain drawing.
func GetBuiltinShader(name ShaderName) (*ebiten.Shader, error) {
	if s, ok := shaderCache[name]; ok {
		return s, nil
	}

	src, ok := builtinSources[name]
	if !ok {
		return nil, fmt.Errorf("ecs: unknown built-in shader %q", name)
	}

	s, err := ebiten.NewShader([]byte(src))
	if err != nil {
		return nil, fmt.Errorf("ecs: compile shader %q: %w", name, err)
	}
	shaderCache[name] = s
	return s, nil
}

// LoadShader compiles an arbitrary Kage source string.
// On failure it returns (nil, error); callers must handle the nil gracefully
// (Placeholder Rule from SKILL.md).
func LoadShader(kageSource string) (*ebiten.Shader, error) {
	s, err := ebiten.NewShader([]byte(kageSource))
	if err != nil {
		return nil, fmt.Errorf("ecs: compile custom shader: %w", err)
	}
	return s, nil
}

// builtinSources maps names to raw Kage source code.
var builtinSources = map[ShaderName]string{
	ShaderNameGrayscale: ShaderGrayscale,
	ShaderNameSepia:     ShaderSepia,
	ShaderNameFadeBlack: ShaderFadeBlack,
}
