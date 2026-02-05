package main

import (
	"context"
	"fmt"
	"os"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	texttospeechpb "cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"google.golang.org/api/option"
)

// ttsClient wraps the Google Cloud TTS client.
type ttsClient struct {
	inner *texttospeech.Client
}

func newTTSClient(ctx context.Context) (*ttsClient, error) {
	var opts []option.ClientOption
	if project := os.Getenv("GOOGLE_CLOUD_PROJECT"); project != "" {
		opts = append(opts, option.WithQuotaProject(project))
	}
	c, err := texttospeech.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("creating TTS client: %w", err)
	}
	return &ttsClient{inner: c}, nil
}

func (c *ttsClient) Close() error {
	return c.inner.Close()
}

// Synthesize converts SSML text to MP3 audio bytes.
func (c *ttsClient) Synthesize(ctx context.Context, ssml string) ([]byte, error) {
	resp, err := c.inner.SynthesizeSpeech(ctx, &texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Ssml{Ssml: ssml},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: "en-US",
			Name:         "en-US-Wavenet-F",
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("synthesizing speech: %w", err)
	}
	return resp.AudioContent, nil
}
