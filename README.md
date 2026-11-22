# Edeth: Audio Watermark Remover & Media Tool

**Edeth** is a web-based media processing tool built with **Go** and **FFmpeg**.

Its flagship feature is a custom audio engine that automatically finds and removes repeating audio watermarks (like demo chimes or vocal tags) from music tracks in under two seconds. Instead of relying on heavy AI models, it uses pure math (FFT and Spectral Subtraction) directly in memory.

It also provides a simple, modern web interface for everyday video and audio editing tasks, saving you from memorizing complex FFmpeg terminal commands.
<img width="1918" height="907" alt="image_2026-11-22_20-32-19" src="https://github.com/user-attachments/assets/edfa2500-533a-429c-b558-51da635daff8" />
![Go](https://img.shields.io/badge/Go-1.27-00ADD8?logo=go)
![FFmpeg](https://img.shields.io/badge/FFmpeg-9.0-green)
![License](https://img.shields.io/badge/license-MIT-blue)


## ✨ Features

* **🎙️ Audio Watermark Removal:** Fast and accurate removal of repeating audio tags using FFT Cross-Correlation and Spectral Subtraction (No AI/GPU required).

* **✂️ Video Trimming:** Supports both blazing-fast cuts (stream-copy) and frame-accurate cutting.

* **🎵 Audio Extraction:** Easily extract and convert video audio to MP3, WAV, or FLAC.

* **📝 Subtitles:** Embed soft subtitles (`.srt`) or burn hard subtitles directly into the video.

* **🌐 Web GUI:** A clean, responsive interface to manage your files and processing tasks.

* **🐳 Docker Ready:** Get up and running in seconds with a fully containerized environment.

## 🚀 Quick Start (Recommended: Docker)

The easiest way to run Edeth without worrying about dependencies is via Docker.

### Prerequisites

* [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/)

### Execution

1. **Clone the repository:**

   ```bash
   git clone https://github.com/yourusername/edeth.git
   cd edeth
   ```

2. **Build and start the container:**

   ```bash
   docker-compose up -d --build
   ```

3. **Open your browser** and go to: `http://localhost:8080`

## 💻 Manual Setup (Local Environment)

If you prefer to run it directly on your machine without Docker:

### Prerequisites

* **Go (1.20+)**: [Download Go](https://go.dev/doc/install)

* **FFmpeg (6.0+)**: [Download FFmpeg](https://ffmpeg.org/download.html) (Make sure it's in your system `PATH`)

### Build & Run

1. **Clone the repo and navigate to it:**

   ```bash
   git clone https://github.com/yourusername/edeth.git
   cd edeth
   ```

2. **Install Go dependencies:**

   ```bash
   go mod tidy
   ```

3. **Run the application:**

   ```bash
   go run main.go server.go handlers.go ffmpeg.go watermark.go
   ```

4. **Access the app** at `http://localhost:8080`.

## 🛠️ How It Works (Under the Hood)

Edeth is designed for speed. It uses Go's `os/exec` and `io` packages to stream raw audio bytes directly from FFmpeg into RAM (Memory Pipes). This means the Digital Signal Processing (DSP) core—which handles the Fourier Transforms (FFT) and phase cancellation—runs directly in memory without creating temporary files on your hard drive.

### Watermark Removal Pipeline:

1. **Decode** both the watermarked audio and watermark sample to raw PCM (mono).

2. **Cross-correlate** to find exact positions where the watermark appears.

3. **Compute optimal gain** at each position for volume alignment.

4. **Phase invert** the watermark and overlay at detected positions.

5. **Mix** the cancellation track with the original to remove the watermark.

6. **Encode** the clean audio to the desired output format.

## 🗂️ Project Structure

```text
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

## 🔌 API Endpoints

| Method | Endpoint | Description | 
| ----- | ----- | ----- | 
| `GET` | `/api/drives` | List available drives | 
| `GET` | `/api/browse?path=...` | Browse directory contents | 
| `GET` | `/api/media-info?path=...` | Get media file metadata | 
| `GET` | `/api/stream?path=...` | HTTP Range media stream & on-the-fly .srm to WAV transcode | 
| `POST` | `/api/upload` | High-throughput streaming upload directly to disk (O(1) RAM) | 
| `POST` | `/api/watermark/remove` | Remove audio watermark with FFT & phase cancellation | 
| `POST` | `/api/video/compress` | Smart constrained bitrate & CRF video compressor | 
| `POST` | `/api/subtitle/add` | Add subtitles to video | 
| `POST` | `/api/trim` | Trim audio/video | 
| `POST` | `/api/audio/extract` | Extract audio from video | 
| `POST` | `/api/audio/merge` | Merge audio into video | 
| `GET` | `/api/progress/{taskId}` | Real-time SSE progress & live stderr terminal stream | 

## 🛠️ Tech Stack

* **Backend**: Pure Go standard library (zero external dependencies)

* **Frontend**: Vanilla HTML/CSS/JS (no frameworks)

* **Media Engine**: FFmpeg via CLI

* **Progress Tracking**: Server-Sent Events (SSE)

* **Embedding**: Go `embed` package for single-binary distribution

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
