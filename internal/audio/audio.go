package audio

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"path"

	eaudio "github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const sampleRate = 48000

// Manager handles BGM and SE playback with graceful fallback.
type Manager struct {
	ctx          *eaudio.Context
	fsys         fs.FS
	bgmPlayer    *eaudio.Player
	bgmVolume    float64
	seVolume     float64
	MasterVolume float64
}

// NewManager creates an audio manager.
func NewManager(fsys fs.FS) *Manager {
	return &Manager{
		ctx:          eaudio.NewContext(sampleRate),
		fsys:         fsys,
		bgmVolume:    0.7,
		seVolume:     0.8,
		MasterVolume: 1.0,
	}
}

// PlayBGM starts looping background music. Stops any current BGM.
// Silently fails if the file is missing (SKILL.md: audio fallback).
func (m *Manager) PlayBGM(name string) {
	m.StopBGM()

	stream, length, err := m.decodeAudio(name)
	if err != nil {
		log.Printf("[WARN] audio: BGM %q failed: %v — silent", name, err)
		return
	}

	loop := eaudio.NewInfiniteLoop(stream, length)
	player, err := m.ctx.NewPlayer(loop)
	if err != nil {
		log.Printf("[WARN] audio: BGM player creation failed: %v", err)
		return
	}

	player.SetVolume(m.bgmVolume * m.MasterVolume)
	player.Play()
	m.bgmPlayer = player
	log.Printf("[INFO] audio: playing BGM %q", name)
}

// StopBGM stops the current background music.
func (m *Manager) StopBGM() {
	if m.bgmPlayer != nil {
		m.bgmPlayer.Close()
		m.bgmPlayer = nil
	}
}

// PlaySE plays a one-shot sound effect.
func (m *Manager) PlaySE(name string) {
	stream, _, err := m.decodeAudio(name)
	if err != nil {
		log.Printf("[WARN] audio: SE %q failed: %v — silent", name, err)
		return
	}

	player, err := m.ctx.NewPlayer(stream)
	if err != nil {
		log.Printf("[WARN] audio: SE player creation failed: %v", err)
		return
	}

	player.SetVolume(m.seVolume * m.MasterVolume)
	player.Play()
	// SE players are fire-and-forget; GC will collect them after playback
}

// SetMasterVolume sets global volume multiplier.
func (m *Manager) UpdateVolume() {
	if m.bgmPlayer != nil {
		m.bgmPlayer.SetVolume(m.bgmVolume * m.MasterVolume)
	}
}

// SetBGMVolume sets BGM volume (0.0 ~ 1.0).
func (m *Manager) SetBGMVolume(v float64) {
	m.bgmVolume = clamp(v)
	m.UpdateVolume()
}

// SetSEVolume sets SE volume (0.0 ~ 1.0).
func (m *Manager) SetSEVolume(v float64) {
	m.seVolume = clamp(v)
}

// decodeAudio reads and decodes an audio file by extension (.ogg, .wav, .mp3).
func (m *Manager) decodeAudio(name string) (io.ReadSeeker, int64, error) {
	clean := path.Clean(name)
	f, err := m.fsys.Open(clean)
	if err != nil {
		return nil, 0, err
	}

	rs, ok := f.(io.ReadSeeker)
	if !ok {
		f.Close()
		return nil, 0, fmt.Errorf("audio: %s does not support seeking", clean)
	}

	ext := path.Ext(clean)
	switch ext {
	case ".ogg":
		stream, err := vorbis.DecodeWithSampleRate(sampleRate, rs)
		if err != nil {
			return nil, 0, err
		}
		return stream, stream.Length(), nil
	case ".wav":
		stream, err := wav.DecodeWithSampleRate(sampleRate, rs)
		if err != nil {
			return nil, 0, err
		}
		return stream, stream.Length(), nil
	case ".mp3":
		stream, err := mp3.DecodeWithSampleRate(sampleRate, rs)
		if err != nil {
			return nil, 0, err
		}
		return stream, stream.Length(), nil
	default:
		return nil, 0, fmt.Errorf("audio: unsupported format %q", ext)
	}
}

func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
