# Edeth — Multimedia Editor

A modern multimedia editing tool built with **Go** and **FFmpeg**, featuring a beautiful dark-themed web UI with a green & black color palette.

![Go](https://img.shields.io/badge/Go-1.27-00ADD8?logo=go)
![FFmpeg](https://img.shields.io/badge/FFmpeg-9.0-green)
![License](https://img.shields.io/badge/license-MIT-blue)

## Features

### ⭐ Audio Watermark Remover (Main Feature)
Remove audio watermarks using intelligent cross-correlation analysis and phase inversion.
- Provide the watermarked audio file and a raw watermark sample
- Automatic watermark position detection via cross-correlation
- Volume-aware phase cancellation with optimal gain computation
- Supports MP3, WAV, FLAC, AAC, OGG, and more

### 📝 Add Subtitles
Add subtitle tracks to video files with two modes:
- **Burn-in**: Permanently render subtitles into the video
- **Embed**: Add as a toggleable subtitle stream
- Supports SRT, ASS, SSA, VTT formats

### ✂️ Trim Media
Cut audio or video files with precise start and end timestamps using stream copy (no re-encoding for speed).

### 🎵 Extract Audio
Separate audio tracks from video files into your preferred format:
- MP3, WAV, AAC, FLAC output options

### 🔀 Merge Audio
Add audio tracks to video files with flexible options:
- **Loop mode**: Audio loops to match video duration
- **Timed placement**: Start audio at any timestamp
- Mixes with existing audio track (or replaces if none exists)

### 🗜️ Smart Video Compressor & Web Optimizer (New)
Target file size encoding and web optimization:
- **Discord 25MB preset**: Constrained bitrate encoding tailored to stay right under Discord's 25MB attachment ceiling
- **Telegram 50MB preset**: High quality compression optimized for Telegram web sharing
- **Web 720p / 1080p presets**: High-efficiency CRF streaming standard
- **Custom Size**: Enter any target megabytes directly
- **Codecs**: H.264 (Universal) and H.265 / HEVC (30-40% smaller)
- **Resolution scaling**: 1080p, 720p, 480p, 360p or original

### 🎧 Waveform Visualizer & Interactive A/B Comparison Player (New)
- **Zero-latency A/B Switch**: Toggle instantaneously between Original (Watermarked) and Cleaned Audio during live playback
- **Interactive Canvas Waveform**: Visual timeline with highlighted markers on detected watermark timestamps
- **Media Previews Everywhere**: Embedded audio and video preview players across all feature pages (Watermark, Trim, Extract, Merge, Subtitle, Compress)
- **On-the-fly `.srm` decoding**: Streams custom binary planar SRM watermarks directly into browser-compatible WAV audio via HTTP Range requests

### 📡 Real-Time SSE Logging & Terminal Console (New)
- Every background operation (FFmpeg transcode and DSP audio processing) pipes live stderr and status messages via Server-Sent Events (SSE) into a built-in terminal console in the UI.

### 📂 Built-in File Browser
Browse your filesystem with a beautiful modal file browser:
- Drive listing (Windows)
- Quick access to common folders
- File type filtering by extension
- Color-coded icons for audio, video, subtitle, and other files

## Requirements

- **Go 1.21+** (if building from source)
- **FFmpeg** in PATH (if running natively)
- **Docker** & **Docker Compose** (optional, for zero-dependency containerized run)

## Quick Start with Docker (Recommended)

Run on any OS (Windows, Linux, macOS) without installing Go or FFmpeg manually:

```bash
docker compose up --build -d
```

Open your browser at `http://localhost:8420`.

## Native Build & Run

```bash
# Build binary
go build -o edeth.exe .

# Run
./edeth.exe
```

The app opens automatically in your default browser at `http://localhost:8420`.

## Project Structure

```
edeth/
├── main.go              # Entry point, embeds frontend, starts server
├── server.go            # HTTP server, SSE progress, task management
├── handlers.go          # REST API handlers for all features
├── ffmpeg.go            # FFmpeg/ffprobe wrapper with progress parsing
├── watermark.go         # Watermark removal algorithm
├── go.mod               # Go module definition
├── frontend/
│   ├── index.html       # SPA shell with sidebar navigation
│   ├── css/
│   │   └── style.css    # Complete dark theme styling
│   └── js/
│       ├── components.js    # Reusable UI components
│       ├── filebrowser.js   # File browser modal
│       └── app.js           # Router, page renderers, API calls
└── README.md
```

## How Watermark Removal Works

1. **Decode** both the watermarked audio and watermark sample to raw PCM (mono)
2. **Cross-correlate** to find exact positions where the watermark appears
3. **Compute optimal gain** at each position for volume alignment
4. **Phase invert** the watermark and overlay at detected positions
5. **Mix** the cancellation track with the original to remove the watermark
6. **Encode** the clean audio to the desired output format

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/drives` | List available drives |
| GET | `/api/browse?path=...` | Browse directory contents |
| GET | `/api/media-info?path=...` | Get media file metadata |
| GET | `/api/stream?path=...` | HTTP Range media stream & on-the-fly .srm to WAV transcode |
| POST | `/api/upload` | High-throughput streaming upload directly to disk (O(1) RAM) |
| POST | `/api/watermark/remove` | Remove audio watermark with FFT & phase cancellation |
| POST | `/api/video/compress` | Smart constrained bitrate & CRF video compressor |
| POST | `/api/subtitle/add` | Add subtitles to video |
| POST | `/api/trim` | Trim audio/video |
| POST | `/api/audio/extract` | Extract audio from video |
| POST | `/api/audio/merge` | Merge audio into video |
| GET | `/api/progress/{taskId}` | Real-time SSE progress & live stderr terminal stream |

## Tech Stack

- **Backend**: Pure Go standard library (zero external dependencies)
- **Frontend**: Vanilla HTML/CSS/JS (no frameworks)
- **Media Engine**: FFmpeg via CLI
- **Progress**: Server-Sent Events (SSE)
- **Embedding**: Go `embed` package for single-binary distribution
