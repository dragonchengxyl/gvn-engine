// Package engine — headless.go
// Thin wrappers that delegate to internal/headless for backwards compatibility.
// The actual implementation lives in internal/headless (zero ebiten dependency).
package engine

import "gvn-engine/internal/headless"

// HeadlessResult is an alias kept for callers that reference engine.HeadlessResult.
type HeadlessResult = headless.Result

// RunHeadlessJSON delegates to headless.RunJSON.
func RunHeadlessJSON(name string, data []byte) (*HeadlessResult, error) {
	return headless.RunJSON(name, data)
}

// RunHeadlessNVN delegates to headless.RunNVN.
func RunHeadlessNVN(name, src string) (*HeadlessResult, error) {
	return headless.RunNVN(name, src)
}

// RunHeadlessAuto delegates to headless.RunAuto.
func RunHeadlessAuto(path string, data []byte) (*HeadlessResult, error) {
	return headless.RunAuto(path, data)
}
