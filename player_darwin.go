//go:build darwin

package main

import "os/exec"

// findPlayer returns the best available audio player on macOS.
// afplay is built into macOS and handles MP3 natively.
func findPlayer() (string, []string) {
	players := []struct {
		name string
		args []string
	}{
		{"afplay", nil},
		{"mpv", []string{"--no-video", "--really-quiet"}},
		{"ffplay", []string{"-nodisp", "-autoexit", "-loglevel", "quiet"}},
	}
	for _, p := range players {
		if path, err := exec.LookPath(p.name); err == nil {
			return path, p.args
		}
	}
	return "", nil
}
