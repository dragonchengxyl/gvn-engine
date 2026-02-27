package ecs

import (
	"math"
	"testing"
)

// ── World & Entity tests ────────────────────────────────────────────────────

func TestWorldCreateAndGet(t *testing.T) {
	w := NewWorld()
	e := w.Create("hero")

	if got := w.Get(e.ID); got != e {
		t.Errorf("Get returned wrong entity")
	}
	if w.Len() != 1 {
		t.Errorf("expected Len()=1, got %d", w.Len())
	}
}

func TestWorldRemove(t *testing.T) {
	w := NewWorld()
	e := w.Create("villain")
	w.Remove(e.ID)

	if w.Get(e.ID) != nil {
		t.Error("entity should be nil after Remove")
	}
	if w.Len() != 0 {
		t.Errorf("expected Len()=0, got %d", w.Len())
	}
}

func TestWorldFindByTag(t *testing.T) {
	w := NewWorld()
	w.Create("bg")
	w.Create("char_sakura")

	e := w.FindByTag("char_sakura")
	if e == nil {
		t.Fatal("FindByTag returned nil")
	}
	if e.Tag != "char_sakura" {
		t.Errorf("expected tag 'char_sakura', got %q", e.Tag)
	}
}

func TestWorldEach(t *testing.T) {
	w := NewWorld()
	for i := 0; i < 5; i++ {
		w.Create("")
	}
	count := 0
	w.Each(func(_ *Entity) { count++ })
	if count != 5 {
		t.Errorf("expected 5 iterations, got %d", count)
	}
}

// ── Component tests ──────────────────────────────────────────────────────────

func TestNewTransform(t *testing.T) {
	tr := NewTransform(100, 200)
	if tr.X != 100 || tr.Y != 200 {
		t.Errorf("position mismatch: got (%v, %v)", tr.X, tr.Y)
	}
	if tr.ScaleX != 1.0 || tr.ScaleY != 1.0 {
		t.Error("expected unit scale")
	}
}

// ── TweenSystem tests ────────────────────────────────────────────────────────

func TestTweenSystemLinear(t *testing.T) {
	w := NewWorld()
	e := w.Create("fx")

	var current float64
	e.Tween = &TweenComp{
		From:     0.0,
		To:       1.0,
		Duration: 1.0,
		OnUpdate: func(v float64) { current = v },
	}

	// Advance 0.5 s → expect ~0.5
	TweenSystem(w, 0.5)
	if math.Abs(current-0.5) > 0.001 {
		t.Errorf("at 0.5s: expected 0.5, got %v", current)
	}

	// Advance another 0.5 s → expect 1.0 and Done
	TweenSystem(w, 0.5)
	if !e.Tween.Done {
		t.Error("tween should be Done after full duration")
	}
	if math.Abs(current-1.0) > 0.001 {
		t.Errorf("at end: expected 1.0, got %v", current)
	}
}

func TestTweenSystemStopsWhenDone(t *testing.T) {
	w := NewWorld()
	e := w.Create("fx")

	calls := 0
	e.Tween = &TweenComp{
		From:     0.0,
		To:       1.0,
		Duration: 0.1,
		OnUpdate: func(_ float64) { calls++ },
	}

	TweenSystem(w, 1.0) // completes in one big step
	callsAfterDone := calls
	TweenSystem(w, 1.0) // should NOT call OnUpdate again
	TweenSystem(w, 1.0)

	if calls != callsAfterDone {
		t.Errorf("OnUpdate called %d times after Done; expected %d", calls, callsAfterDone)
	}
}

func TestEasedTween(t *testing.T) {
	var current float64
	tc := EasedTween(0.0, 1.0, 1.0, func(v float64) { current = v })

	w := NewWorld()
	e := w.Create("eased")
	e.Tween = tc

	// At t=0.5 smooth-step ≈ 0.5 (symmetric)
	TweenSystem(w, 0.5)
	if math.Abs(current-0.5) > 0.01 {
		t.Errorf("eased tween at midpoint: expected ~0.5, got %v", current)
	}
}

func TestEntityUniqueIDs(t *testing.T) {
	w := NewWorld()
	ids := make(map[EntityID]bool)
	for i := 0; i < 100; i++ {
		e := w.Create("")
		if ids[e.ID] {
			t.Fatalf("duplicate EntityID: %d", e.ID)
		}
		ids[e.ID] = true
	}
}
