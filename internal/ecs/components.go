package ecs

// Transform holds 2-D position and scale for a rendered entity.
// All values are in virtual screen coordinates (1920×1080).
type Transform struct {
	X, Y           float64
	ScaleX, ScaleY float64 // 1.0 = original size
}

// NewTransform returns a Transform at (x,y) with unit scale.
func NewTransform(x, y float64) *Transform {
	return &Transform{X: x, Y: y, ScaleX: 1.0, ScaleY: 1.0}
}

// Sprite references a GPU image for rendering.
// We store it as an interface so that this file has no direct ebiten import —
// the engine assigns the concrete *ebiten.Image value.
type Sprite struct {
	// Image is the renderable asset. Typed as any to keep this file ebiten-free;
	// the render layer casts it to *ebiten.Image.
	Image any
}

// AlphaComp stores the current transparency of an entity (0.0 opaque → 0.0 invisible).
// Note: 1.0 = fully visible, 0.0 = fully transparent.
type AlphaComp struct {
	Value float64 // clamped [0.0, 1.0]
}

// TweenComp drives a linear animation of a single float64 value over time.
// When Done is true the system stops updating it.
type TweenComp struct {
	From     float64
	To       float64
	Duration float64 // total duration in seconds
	Elapsed  float64 // time elapsed so far
	Done     bool
	// OnUpdate is called each tick with the current interpolated value.
	// Point it at the field you want to animate (e.g. &entity.Alpha.Value).
	OnUpdate func(val float64)
}

// ShaderFXComp attaches a compiled Kage shader and its uniforms to an entity.
// The render layer uses this when drawing the entity's Sprite.
type ShaderFXComp struct {
	// ShaderSrc is the compiled *ebiten.Shader, stored as any for the same
	// ebiten-free reason as Sprite.Image.
	ShaderSrc any
	Uniforms  map[string]any
}
