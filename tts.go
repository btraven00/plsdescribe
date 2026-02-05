package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	spaceCollapser = regexp.MustCompile(`\s+`)
	markdownBullet = regexp.MustCompile(`(?m)^\s*[\*\-•]\s*`)
	markdownBold   = regexp.MustCompile(`\*{1,3}([^*]+)\*{1,3}`)
	markdownHeader = regexp.MustCompile(`(?m)^#{1,6}\s*`)
)

// cleanForTTS strips markdown artifacts and normalizes whitespace so the
// TTS engine doesn't read out "asterisk" or other markup literally.
// List items get a trailing period so the TTS engine pauses between them.
// If the text already contains <speak> tags (SSML from Gemini), they are
// preserved. Otherwise the text is wrapped in <speak> tags.
func cleanForTTS(text string) string {
	// Strip bold/italic markers
	text = markdownBold.ReplaceAllString(text, "$1")
	text = markdownHeader.ReplaceAllString(text, "")

	// Turn each bullet item into a sentence: strip the bullet marker and
	// ensure the line ends with a period so TTS inserts a natural pause.
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		line = markdownBullet.ReplaceAllString(line, "")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Add a period if the line doesn't already end with punctuation.
		last := line[len(line)-1]
		if last != '.' && last != '!' && last != '?' && last != ',' && last != ';' {
			line += "."
		}
		lines = append(lines, line)
	}
	text = strings.Join(lines, " ")

	// Collapse whitespace
	text = spaceCollapser.ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	// Ensure wrapped in <speak> tags for SSML input
	if !strings.Contains(text, "<speak>") {
		text = "<speak>" + text + "</speak>"
	}

	return text
}

func speakText(text string) error {
	ssml := cleanForTTS(text)

	ctx := context.Background()
	client, err := newTTSClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()


	mp3Data, err := client.Synthesize(ctx, ssml)
	if err != nil {
		return err
	}

	// Write MP3 to a temp file and play it.
	tmpFile, err := os.CreateTemp("", "plsdescribe-*.mp3")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write(mp3Data); err != nil {
		tmpFile.Close()
		return fmt.Errorf("writing audio: %w", err)
	}
	tmpFile.Close()

	playerPath, playerArgs := findPlayer()
	if playerPath == "" {
		outPath := filepath.Join(".", "description.mp3")
		if err := os.WriteFile(outPath, mp3Data, 0644); err != nil {
			return fmt.Errorf("saving MP3: %w", err)
		}
		fmt.Fprintf(os.Stderr, "No audio player found. Saved MP3 to %s\n", outPath)
		return nil
	}

	args := append(playerArgs, tmpPath)
	cmd := exec.Command(playerPath, args...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("playing audio with %s: %w", filepath.Base(playerPath), err)
	}
	return nil
}
