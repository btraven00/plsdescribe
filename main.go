package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/genai"
)

const (
	// TODO: switch to "gemini-3-pro" when available in the API
	model = "gemini-2.5-pro"
	field = "bioinformatics"

	promptBase = "You are an assistant to a data scientist, in the field of " + field + ". " +
		"Your task is to describe plots, with minimal interpretation, unless explicitely asked otherwise. " +
		"The goal is to enable accesibility features in data analysis tools. "

	defaultContext = "Additional context: the plot represents an UMAP embedding of different clusters for cell types."

	promptV1Suffix = "Describe this plot in one clear and concise sentence."
	promptV2Suffix = "Describe the key characteristics of the clusters in this plot, focusing on their relative positions, sizes, and separation. " +
		"Use four or less bullet points for your description. " +
		"Enclose answer in <speak> tags, and use basic SSML tags to improve generation, but avoid html tags and <break> in particular."
)

func mimeFromPath(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".bmp":
		return "image/bmp"
	default:
		return "application/octet-stream"
	}
}

// describe sends a prompt + image to Gemini and returns the response text.
func describe(ctx context.Context, client *genai.Client, prompt string, imgPart *genai.Part) (string, error) {
	parts := []*genai.Part{{Text: prompt}}
	if imgPart != nil {
		parts = append(parts, imgPart)
	}

	resp, err := client.Models.GenerateContent(ctx, model, []*genai.Content{
		{Parts: parts},
	}, nil)
	if err != nil {
		return "", err
	}

	var text string
	if resp != nil && len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		for _, part := range resp.Candidates[0].Content.Parts {
			if part.Text != "" {
				text += part.Text
			}
		}
	}
	if text == "" {
		return "", fmt.Errorf("no response generated")
	}
	return text, nil
}

func main() {
	imagePath := flag.String("f", "", "Image file to describe (required)")
	verbose := flag.Bool("v", false, "Increase verbosity (detailed bullet points)")
	question := flag.String("q", "", "A question to append")
	outputFile := flag.String("o", "description.txt", "Output file for the description")
	tts := flag.Bool("tts", false, "Speak the description aloud via Google Cloud TTS")
	interactive := flag.Bool("i", false, "Enter interactive session for follow-up questions")

	flag.Parse()

	if *imagePath == "" {
		fmt.Fprintln(os.Stderr, "Error: -f (image file) is required.")
		flag.Usage()
		os.Exit(1)
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: GEMINI_API_KEY environment variable not set.")
		os.Exit(1)
	}

	imgData, err := os.ReadFile(*imagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading image: %v\n", err)
		os.Exit(1)
	}

	// Image part is built once and reused for every request in the session.
	imgPart := &genai.Part{
		InlineData: &genai.Blob{
			MIMEType: mimeFromPath(*imagePath),
			Data:     imgData,
		},
	}

	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Gemini client: %v\n", err)
		os.Exit(1)
	}

	// Build initial prompt.
	var prompt string
	if *verbose {
		prompt = promptBase + promptV2Suffix + " " + defaultContext
	} else {
		prompt = promptBase + promptV1Suffix + " " + defaultContext
	}
	if *question != "" {
		prompt += " " + *question
	}

	// --- Initial description ---
	description, err := describe(ctx, client, prompt, imgPart)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// When TTS is requested, don't print to stdout — avoids the screen
	// reader speaking over the Wavenet voice.
	if *tts {
		if err := speakText(description); err != nil {
			fmt.Fprintf(os.Stderr, "TTS error: %v\n", err)
		}
	} else {
		fmt.Println(description)
	}

	if *outputFile != "" {
		if err := os.WriteFile(*outputFile, []byte(description), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		}
	}

	// --- Interactive session ---
	if !*interactive {
		return
	}

	lastResponse := description
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Fprintln(os.Stderr, "\nInteractive mode. Commands: /tts, /save [file], /quit")
	fmt.Fprint(os.Stderr, "> ")

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			fmt.Fprint(os.Stderr, "> ")
			continue
		}

		switch {
		case line == "/quit" || line == "/q" || line == "/exit":
			return

		case line == "/tts":
			// Speak the last response instead of printing it.
			if err := speakText(lastResponse); err != nil {
				fmt.Fprintf(os.Stderr, "TTS error: %v\n", err)
			}

		case strings.HasPrefix(line, "/save"):
			fname := strings.TrimSpace(strings.TrimPrefix(line, "/save"))
			if fname == "" {
				fname = *outputFile
			}
			if err := os.WriteFile(fname, []byte(lastResponse), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "Saved to %s\n", fname)
			}

		case line == "/help" || line == "/?":
			fmt.Fprintln(os.Stderr, "Commands:")
			fmt.Fprintln(os.Stderr, "  /tts              Speak the last response (Google Cloud TTS)")
			fmt.Fprintln(os.Stderr, "  <question> /tts   Ask a question and speak the answer")
			fmt.Fprintln(os.Stderr, "  /save [file]      Save last response to file")
			fmt.Fprintln(os.Stderr, "  /quit             Exit interactive mode")
			fmt.Fprintln(os.Stderr, "  /help             Show this help")
			fmt.Fprintln(os.Stderr, "Anything else is sent as a follow-up question about the image.")

		default:
			// If the question ends with /tts, speak the answer instead
			// of printing to stdout (avoids screen reader collision).
			speakAnswer := strings.HasSuffix(line, "/tts")
			if speakAnswer {
				line = strings.TrimSpace(strings.TrimSuffix(line, "/tts"))
			}

			followUp := promptBase + defaultContext +
				" Previous description: " + lastResponse +
				" User question: " + line

			resp, err := describe(ctx, client, followUp, imgPart)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			} else {
				lastResponse = resp
				if speakAnswer {
					if err := speakText(resp); err != nil {
						fmt.Fprintf(os.Stderr, "TTS error: %v\n", err)
					}
				} else {
					fmt.Println(resp)
				}
			}
		}

		fmt.Fprint(os.Stderr, "> ")
	}
}
