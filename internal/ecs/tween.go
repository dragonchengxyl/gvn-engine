package ecs

// TweenSystem advances all active TweenComp components by dt seconds.
// It reads Elapsed/Duration to compute a [0,1] progress, calls OnUpdate with
// the interpolated value, and marks Done when the animation completes.
//
// Call this once per Update tick: TweenSystem(world, 1.0/60.0)
func TweenSystem(w *World, dt float64) {
	w.Each(func(e *Entity) {
		tc := e.Tween
		if tc == nil || tc.Done {
			return
		}
		tc.Elapsed += dt
		t := tc.Elapsed / tc.Duration
		if t >= 1.0 {
			t = 1.0
			tc.Done = true
		}
		val := lerp(tc.From, tc.To, t)
		if tc.OnUpdate != nil {
			tc.OnUpdate(val)
		}
	})
}

// lerp linearly interpolates between a and b by factor t ∈ [0,1].
func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// EasedTween returns a TweenComp that uses a smooth-step (ease-in-out) curve
// instead of linear interpolation.
func EasedTween(from, to, duration float64, onUpdate func(float64)) *TweenComp {
	return &TweenComp{
		From:     from,
		To:       to,
		Duration: duration,
		OnUpdate: func(t float64) {
			// smooth-step: 3t²-2t³
			s := smoothStep(0, 1, t)
			if onUpdate != nil {
				onUpdate(lerp(from, to, s))
			}
		},
	}
}

// smoothStep maps t from [edge0, edge1] to a smooth [0, 1].
func smoothStep(edge0, edge1, t float64) float64 {
	if t <= edge0 {
		return 0
	}
	if t >= edge1 {
		return 1
	}
	x := (t - edge0) / (edge1 - edge0)
	return x * x * (3 - 2*x)
}
