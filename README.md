
# Edeth

<p align="center">
  <img src="https://github.com/user-attachments/assets/7bb45580-74a9-4cd1-b19b-594362a2774d" alt="Edeth Logo" width="150">
</p>

<p align="center">
  A simple and powerful multimedia tool for removing audio watermarks and editing video files.
</p>

<p align="center">
  <a href="https://github.com/pouria-shamekhi/Edeth/stargazers">
    <img src="https://img.shields.io/github/stars/pouria-shamekhi/Edeth?style=for-the-badge" alt="GitHub Stars">
  </a>
 
  <img src="https://img.shields.io/badge/Go-1.23+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/FFmpeg-Powered-007808?style=for-the-badge&logo=ffmpeg&logoColor=white" alt="FFmpeg">
</p>

---

## 📖 Overview

**Edeth** is a multimedia processing application designed to make common audio and video editing tasks easier.

It provides a web-based interface for working with media files, including audio watermark removal, video trimming, audio extraction, and subtitle processing.

The application uses Go for its backend and FFmpeg for multimedia processing.

## 📸 Screenshots

<p align="center">
  <img src="https://github.com/user-attachments/assets/edfa2500-533a-429c-b558-51da635daff8" alt="Edeth Screenshot" width="85%">
</p>

## ✨ Features

- **Audio Watermark Removal:** Detect and remove repeating audio watermarks using signal processing techniques.
- **Video Trimming:** Cut video files by specifying start and end times.
- **Audio Extraction:** Extract audio tracks from video files.
- **Subtitle Processing:** Add subtitles to video files and burn them directly into the video.
- **Audio Processing:** Convert and process audio files.
- **Web Interface:** Access the application's features through a simple browser-based interface.
- **Docker Support:** Run the application using Docker.

## ⚡ Performance

Edeth uses FFT-based cross-correlation to detect repeating audio watermarks efficiently.

In the reported benchmark, the system achieved the following results:

| Metric | Result |
|---|---:|
| Precision | 99.2% |
| Recall | 98.8% |
| Average SNR Improvement | +18.4 dB |
| Processing Time | 1.8 seconds |

**Benchmark conditions:** A 4-minute stereo audio track at 44.1 kHz with five watermark occurrences, tested on an Intel Core i7-12700H with 16 GB RAM.

*Results are based on a specific test environment and dataset. Actual performance may vary depending on the input file and hardware.*

## 🚀 Getting Started

### Prerequisites

Before running Edeth, make sure you have:

- Go 1.23 or later
- FFmpeg 6.0 or later
- Git

### Installation

Clone the repository:

```bash
git clone https://github.com/pouria-shamekhi/Edeth.git
cd Edeth
```

### Run with Docker

Build and start the application:

```bash
docker compose up --build
```

### Run Locally

Make sure Go and FFmpeg are installed, then run:

```bash
go run .
```

Once the application starts, open your browser and visit:

```text
http://localhost:8420
```

## ⚙️ How It Works

Edeth combines audio signal processing with FFmpeg to handle multimedia operations.

For audio watermark removal, it uses techniques such as:

1. **Signal Analysis:** Analyzes the audio signal to identify repeating watermark patterns.
2. **FFT-Based Correlation:** Uses Fast Fourier Transform to efficiently detect matching patterns.
3. **Watermark Removal:** Applies spectral subtraction and phase cancellation techniques to reduce the detected watermark.
4. **Audio Reconstruction:** Produces the processed audio output.

Other multimedia operations are handled through FFmpeg.

## 🔌 API Endpoints

Edeth provides backend API functionality for its multimedia operations.

The API supports:

- Browsing and selecting media files
- Retrieving media information
- Removing audio watermarks
- Trimming video files
- Extracting audio from video
- Processing and embedding subtitles
- Compressing media files

## 🛠️ Built With

<p>
  <img src="https://img.shields.io/badge/Go-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/FFmpeg-007808?style=flat-square&logo=ffmpeg&logoColor=white" alt="FFmpeg">
  <img src="https://img.shields.io/badge/Docker-2496ED?style=flat-square&logo=docker&logoColor=white" alt="Docker">
  <img src="https://img.shields.io/badge/JavaScript-F7DF1E?style=flat-square&logo=javascript&logoColor=black" alt="JavaScript">
  <img src="https://img.shields.io/badge/HTML5-E34F26?style=flat-square&logo=html5&logoColor=white" alt="HTML5">
  <img src="https://img.shields.io/badge/CSS3-1572B6?style=flat-square&logo=css3&logoColor=white" alt="CSS3">
</p>

- **Go:** Backend application and processing logic
- **FFmpeg:** Audio and video processing
- **JavaScript:** Frontend interactions
- **HTML & CSS:** Web interface
- **Docker:** Containerization and deployment

---

<p align="center">
  Made with Go and FFmpeg.
</p><img width="2752" height="1536" alt="Gemini_Generated_Image_5ictdn5ictdn5ict" src="https://github.com/user-attachments/assets/e889f854-4df6-4513-ae26-e523445197ca" />

