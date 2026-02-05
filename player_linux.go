//go:build linux

package main

import "os/exec"

// findPlayer returns the best available audio player on Linux.
// paplay (PulseAudio/PipeWire) and mpv are the most common; pw-play for
// modern PipeWire-native setups. ffplay as a fallback.
func findPlayer() (string, []string) {
	players := []struct {
		name string
		args []string
	}{
		{"mpv", []string{"--no-video", "--really-quiet"}},
		{"pw-play", nil},
		{"paplay", nil},
		{"ffplay", []string{"-nodisp", "-autoexit", "-loglevel", "quiet"}},
	}
	for _, p := range players {
		if path, err := exec.LookPath(p.name); err == nil {
			return path, p.args
		}
	}
	return "", nil
}
