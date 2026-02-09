package main

import (
	"context"
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	texttospeechpb "cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/api/option"
)

type synthesizeRequest struct {
	SSML         string  `json:"ssml"`
	SpeakingRate float64 `json:"speaking_rate"`
}

func main() {
	token := os.Getenv("TTS_PROXY_TOKEN")
	if token == "" {
		log.Fatal("TTS_PROXY_TOKEN must be set")
	}

	domain := os.Getenv("TTS_PROXY_DOMAIN")

	ctx := context.Background()

	// Build Google Cloud TTS client (shared across requests).
	var opts []option.ClientOption
	if project := os.Getenv("GOOGLE_CLOUD_PROJECT"); project != "" {
		opts = append(opts, option.WithQuotaProject(project))
	}
	ttsClient, err := texttospeech.NewClient(ctx, opts...)
	if err != nil {
		log.Fatalf("creating TTS client: %v", err)
	}
	defer ttsClient.Close()

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Health check — no auth.
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	// Authenticated API group.
	v1 := e.Group("/v1")
	v1.Use(tokenAuth(token))

	v1.POST("/synthesize", func(c echo.Context) error {
		var req synthesizeRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
		}
		if req.SSML == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "ssml is required")
		}

		audioCfg := &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
		}
		if req.SpeakingRate > 0 {
			audioCfg.SpeakingRate = req.SpeakingRate
		}

		resp, err := ttsClient.SynthesizeSpeech(c.Request().Context(), &texttospeechpb.SynthesizeSpeechRequest{
			Input: &texttospeechpb.SynthesisInput{
				InputSource: &texttospeechpb.SynthesisInput_Ssml{Ssml: req.SSML},
			},
			Voice: &texttospeechpb.VoiceSelectionParams{
				LanguageCode: "en-US",
				Name:         "en-US-Wavenet-F",
			},
			AudioConfig: audioCfg,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusBadGateway, fmt.Sprintf("TTS error: %v", err))
		}

		return c.Blob(http.StatusOK, "audio/mpeg", resp.AudioContent)
	})

	if domain != "" {
		e.AutoTLSManager.HostPolicy = autocert.HostWhitelist(domain)
		e.AutoTLSManager.Cache = autocert.DirCache("/var/cache/tts-proxy/certs")
		log.Printf("Starting with AutoTLS for %s", domain)
		e.Logger.Fatal(e.StartAutoTLS(":443"))
	} else {
		addr := os.Getenv("TTS_PROXY_ADDR")
		if addr == "" {
			addr = ":8080"
		}
		log.Printf("Starting on %s (no TLS — set TTS_PROXY_DOMAIN for AutoTLS)", addr)
		e.Logger.Fatal(e.Start(addr))
	}
}

// tokenAuth returns middleware that checks for a valid Bearer token.
func tokenAuth(expected string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing bearer token")
			}
			got := strings.TrimPrefix(auth, "Bearer ")
			if subtle.ConstantTimeCompare([]byte(got), []byte(expected)) != 1 {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}
			return next(c)
		}
	}
}
