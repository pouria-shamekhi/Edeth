# Edeth Sample Audio Files

This directory contains test and demo media files for testing the **Audio Watermark Remover** feature:

| File | Description | Format |
|------|-------------|--------|
| `music_watermarked.mp3` | A song with repeating background voice watermarks (~13 occurrences). | MP3 (44.1 kHz, Stereo) |
| `wateraudio.srm` | Custom raw binary planar float32 watermark reference signature. | SRM (DNZ 44.1 kHz, Stereo) |
| `AudioJungle watermark sound.mp3` | Standard audio watermark audio sample. | MP3 (44.1 kHz, Stereo) |

## Quick Testing Instructions
1. Open Edeth at `http://localhost:8420`.
2. Go to **Audio Watermark Remover**.
3. Select `samples/music_watermarked.mp3` as the input audio.
4. Select `samples/wateraudio.srm` (or `samples/AudioJungle watermark sound.mp3`) as the watermark reference.
5. Click **Remove Watermark**.
6. Listen to the real-time **A/B comparison player** to hear the instant difference between the original and clean audio!
