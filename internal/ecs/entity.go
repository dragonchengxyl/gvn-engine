// Package ecs provides a lightweight Entity-Component-System used by GVN-Nexus
// for animated sprites, shader effects, and non-blocking transitions.
//
// Design goals:
//   - Zero reflection: all component access is type-safe via typed struct fields.
//   - No ebiten import in this file (keeps the package headless-friendly).
//   - A World acts as the ECS registry; systems are plain functions operating on it.
package ecs

import "sync/atomic"

// EntityID is a unique, monotonically increasing entity identifier.
type EntityID uint64

var lastID uint64

func newEntityID() EntityID {
	return EntityID(atomic.AddUint64(&lastID, 1))
}

// Entity holds all optional component pointers.
// A nil pointer means "this component is not present".
type Entity struct {
	ID        EntityID
	Transform *Transform
	Sprite    *Sprite
	Alpha     *AlphaComp
	Tween     *TweenComp
	ShaderFX  *ShaderFXComp
	// Tag is an arbitrary string for debugging / lookup by name.
	Tag string
}

// World is the ECS entity registry.
type World struct {
	entities map[EntityID]*Entity
}

// NewWorld creates an empty World.
func NewWorld() *World {
	return &World{entities: make(map[EntityID]*Entity)}
}

// Create registers a new Entity and returns it.
func (w *World) Create(tag string) *Entity {
	e := &Entity{ID: newEntityID(), Tag: tag}
	w.entities[e.ID] = e
	return e
}

// Remove deletes an entity from the world.
func (w *World) Remove(id EntityID) {
	delete(w.entities, id)
}

// Get returns the entity with the given ID, or nil.
func (w *World) Get(id EntityID) *Entity {
	return w.entities[id]
}

// FindByTag returns the first entity whose Tag matches, or nil.
func (w *World) FindByTag(tag string) *Entity {
	for _, e := range w.entities {
		if e.Tag == tag {
			return e
		}
	}
	return nil
}

// Each calls fn for every registered entity.
func (w *World) Each(fn func(*Entity)) {
	for _, e := range w.entities {
		fn(e)
	}
}

// Len returns the number of registered entities.
func (w *World) Len() int { return len(w.entities) }
