# plsdescribe

Describe scientific plots using AI vision, designed for visually impaired researchers.

Sends an image to Google Gemini, gets a text description back, and optionally speaks it aloud via Google Cloud TTS. Descriptions go to stdout so screen readers (VoiceOver, Orca) pick them up naturally.

## Install

Grab a binary from [Releases](../../releases), or build from source:

```
go build -o plsdescribe .
```

## Usage

```
export GEMINI_API_KEY=your-key

plsdescribe -f plot.png                 # concise one-sentence description
plsdescribe -f plot.png -v              # detailed bullet points
plsdescribe -f plot.png -i              # interactive session
plsdescribe -f plot.png -v -tts         # describe and speak aloud
```

### Interactive mode

`-i` opens a session where the image is loaded once and you can ask follow-up questions:

```
$ plsdescribe -f umap.png -i
The UMAP plot shows 21 cell-type clusters distributed across two dimensions...

> how many clusters overlap?
> which cluster is the largest?
> /tts
> /save notes.txt
> /quit
```

Commands: `/tts` (speak last response), `/save [file]` (save to file), `/quit`, `/help`.

### Flags

| Flag | Description |
|------|-------------|
| `-f` | Image file to describe (required) |
| `-v` | Verbose output (detailed bullet points) |
| `-i` | Interactive session for follow-up questions |
| `-q` | Append a question to the initial prompt |
| `-o` | Output file (default: `description.txt`) |
| `-tts` | Speak via Google Cloud TTS |

## TTS setup

TTS requires Google Cloud credentials with the Text-to-Speech API enabled:

```
gcloud auth application-default login
gcloud auth application-default set-quota-project your-project-id
```

Or set `GOOGLE_CLOUD_PROJECT` to your GCP project ID.

Audio plays through `mpv`, `ffplay`, `afplay` (macOS), or `pw-play`/`paplay` (Linux) — whichever is found first. If none are available, the MP3 is saved to `description.mp3`.

## Screen reader compatibility

All descriptions print to stdout; UI prompts and status go to stderr. TTS only fires when explicitly requested (`-tts` flag or `/tts` command), so it never collides with system screen readers.
