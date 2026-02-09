package main

import (
	"context"
	"fmt"
	"os"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	texttospeechpb "cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"google.golang.org/api/option"
)

// synthesizer abstracts TTS backends so the CLI can use either the direct
// Google Cloud client or an HTTP proxy transparently.
type synthesizer interface {
	Synthesize(ctx context.Context, ssml string, rate float64) ([]byte, error)
	Close() error
}

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
// rate controls speaking speed (0.25–2.0); 0 means default (1.0).
func (c *ttsClient) Synthesize(ctx context.Context, ssml string, rate float64) ([]byte, error) {
	audioCfg := &texttospeechpb.AudioConfig{
		AudioEncoding: texttospeechpb.AudioEncoding_MP3,
	}
	if rate > 0 {
		audioCfg.SpeakingRate = rate
	}

	resp, err := c.inner.SynthesizeSpeech(ctx, &texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Ssml{Ssml: ssml},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: "en-US",
			Name:         "en-US-Wavenet-F",
		},
		AudioConfig: audioCfg,
	})
	if err != nil {
		return nil, fmt.Errorf("synthesizing speech: %w", err)
	}
	return resp.AudioContent, nil
}
