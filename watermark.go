package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"math/cmplx"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// WatermarkMatch represents a detected watermark occurrence.
type WatermarkMatch struct {
	Offset      int     `json:"offset"`
	Gain        float64 `json:"gain"`
	Correlation float64 `json:"correlation"`
}

type peakCandidate struct {
	offset int
	corr   float64
	rho    float64
}

// SRMData holds parsed planar float32 audio from an .srm file.
type SRMData struct {
	Channels   int
	SampleRate int
	NumSamples int
	Ch0        []float32
	Ch1        []float32
}

// parseSRMFile parses a custom SRM binary watermark file (DNZ format).
func parseSRMFile(path string) (*SRMData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read SRM file: %v", err)
	}
	if len(data) < 43 || string(data[:3]) != "DNZ" {
		return nil, fmt.Errorf("file does not have DNZ header")
	}

	channels := int(data[11])
	if channels < 1 {
		channels = 1
	}
	sampleRate := int(binary.LittleEndian.Uint32([]byte{data[15], data[16], data[17], data[18]}))
	if sampleRate <= 0 {
		sampleRate = 44100
	}
	numSamples := int(binary.LittleEndian.Uint32(data[39:43]))
	if numSamples <= 0 {
		return nil, fmt.Errorf("invalid sample count in SRM header")
	}

	stride := numSamples * 4
	if 43+stride > len(data) {
		return nil, fmt.Errorf("SRM file truncated: expected %d bytes, got %d", 43+stride, len(data))
	}

	ch0 := make([]float32, numSamples)
	for i := 0; i < numSamples; i++ {
		bits := binary.LittleEndian.Uint32(data[43+i*4 : 43+(i+1)*4])
		ch0[i] = math.Float32frombits(bits)
	}

	var ch1 []float32
	if channels > 1 && 43+stride*2 <= len(data) {
		ch1 = make([]float32, numSamples)
		for i := 0; i < numSamples; i++ {
			bits := binary.LittleEndian.Uint32(data[43+stride+i*4 : 43+stride+(i+1)*4])
			ch1[i] = math.Float32frombits(bits)
		}
	} else {
		ch1 = ch0
	}

	return &SRMData{
		Channels:   channels,
		SampleRate: sampleRate,
		NumSamples: numSamples,
		Ch0:        ch0,
		Ch1:        ch1,
	}, nil
}

// RemoveWatermark detects and removes audio watermarks using FFT cross-correlation,
// exact least-squares phase cancellation, and spectral subtraction.
func RemoveWatermark(inputFile, watermarkFile, outputFile string, task *Task) error {
	inputFile = cleanPath(inputFile)
	watermarkFile = cleanPath(watermarkFile)
	outputFile = cleanPath(outputFile)

	sendTaskUpdate(task, 5, "processing", "Analyzing audio files...")

	const targetSampleRate = 44100

	// 1. Decode main audio (preserve stereo if stereo)
	sendTaskLog(task, 10, "processing", "Decoding main audio...", fmt.Sprintf("Streaming PCM from %s via FFmpeg...", filepath.Base(inputFile)))
	mainAudio, mainChannels, err := decodeAudioFile(inputFile, targetSampleRate, false)
	if err != nil {
		return fmt.Errorf("failed to decode main audio: %v", err)
	}
	if len(mainAudio) == 0 {
		return fmt.Errorf("main audio contains no samples")
	}

	// 2. Decode watermark audio (preserve channels so stereo cancels stereo)
	sendTaskLog(task, 20, "processing", "Decoding watermark reference...", fmt.Sprintf("Parsing reference watermark from %s...", filepath.Base(watermarkFile)))
	wmAudio, wmChannels, err := decodeAudioFile(watermarkFile, targetSampleRate, false)
	if err != nil {
		return fmt.Errorf("failed to decode watermark: %v", err)
	}
	if len(wmAudio) == 0 {
		return fmt.Errorf("watermark audio contains no samples")
	}

	numSongFrames := len(mainAudio) / mainChannels
	numWmFrames := len(wmAudio) / wmChannels

	// Extract watermark channel slices
	rawWmL := make([]float64, numWmFrames)
	rawWmR := make([]float64, numWmFrames)
	for i := 0; i < numWmFrames; i++ {
		lVal := float64(wmAudio[i*wmChannels])
		rVal := lVal
		if wmChannels > 1 {
			rVal = float64(wmAudio[i*wmChannels+1])
		}
		rawWmL[i] = lVal
		rawWmR[i] = rVal
	}

	// Trim leading and trailing silence from the watermark template (< 0.0001 amplitude)
	// This ensures correlation aligns with the actual watermark waveform rather than padding silence
	silenceStart := 0
	silenceEnd := numWmFrames - 1
	for silenceStart < numWmFrames && math.Abs(rawWmL[silenceStart]) < 0.0001 && math.Abs(rawWmR[silenceStart]) < 0.0001 {
		silenceStart++
	}
	for silenceEnd > silenceStart && math.Abs(rawWmL[silenceEnd]) < 0.0001 && math.Abs(rawWmR[silenceEnd]) < 0.0001 {
		silenceEnd--
	}
	if silenceEnd <= silenceStart {
		// Watermark is completely silent
		silenceStart = 0
		silenceEnd = numWmFrames - 1
	}

	trimmedWmFrames := silenceEnd - silenceStart + 1
	wmL := make([]float64, trimmedWmFrames)
	wmR := make([]float64, trimmedWmFrames)
	monoWM := make([]float64, trimmedWmFrames)
	for i := 0; i < trimmedWmFrames; i++ {
		wmL[i] = rawWmL[silenceStart+i]
		wmR[i] = rawWmR[silenceStart+i]
		monoWM[i] = 0.5 * (wmL[i] + wmR[i])
	}

	if trimmedWmFrames > numSongFrames {
		return fmt.Errorf("watermark duration (%d samples) is longer than audio (%d samples)", trimmedWmFrames, numSongFrames)
	}

	// 3. Create mono search template from song
	sendTaskUpdate(task, 25, "processing", "Preparing FFT correlation vectors...")
	monoSong := make([]float64, numSongFrames)
	for i := 0; i < numSongFrames; i++ {
		lVal := float64(mainAudio[i*mainChannels])
		rVal := lVal
		if mainChannels > 1 {
			rVal = float64(mainAudio[i*mainChannels+1])
		}
		monoSong[i] = 0.5 * (lVal + rVal)
	}

	// 4. Find all watermark occurrences using FFT Cross-Correlation
	sendTaskLog(task, 35, "processing", "Searching for watermark with FFT cross-correlation...", fmt.Sprintf("Computing FFT cross-correlation over %d audio frames...", numSongFrames))
	matches := detectAllWatermarkPeaks(monoSong, monoWM, targetSampleRate)

	sendTaskLog(task, 55, "processing", fmt.Sprintf("Found %d watermark occurrence(s)", len(matches)), fmt.Sprintf("Correlation completed: %d watermark occurrences detected above threshold", len(matches)))

	if len(matches) == 0 {
		sendTaskLog(task, 75, "processing", "No watermark detected above threshold, copying original...", "Zero peaks found above detection threshold. Preserving original file.")
		return encodeFloat32Audio(mainAudio, targetSampleRate, mainChannels, outputFile, func(p float64) {
			sendTaskLog(task, 75+p*0.24, "processing", "Encoding output...", fmt.Sprintf("Encoding clean audio: %.1f%%", p))
		})
	}

	// 5. Channel-specific Phase Cancellation & Spectral Subtraction (Equations 8-11 in paper)
	sendTaskLog(task, 65, "processing", fmt.Sprintf("Removing %d watermark instance(s)...", len(matches)), "Starting dual-stage time-domain phase cancellation and spectral subtraction...")

	// De-interleave song into L and R channels
	songL := make([]float64, numSongFrames)
	songR := make([]float64, numSongFrames)
	for i := 0; i < numSongFrames; i++ {
		songL[i] = float64(mainAudio[i*mainChannels])
		if mainChannels > 1 {
			songR[i] = float64(mainAudio[i*mainChannels+1])
		} else {
			songR[i] = songL[i]
		}
	}

	// Precompute watermark channel energies
	var wmEL, wmER float64
	for i := 0; i < trimmedWmFrames; i++ {
		wmEL += wmL[i] * wmL[i]
		wmER += wmR[i] * wmR[i]
	}

	const fadeLen = 64

	for idx, m := range matches {
		nominalOffset := m.Offset

		// Fine alignment within +/- 16 samples to minimize least-squares residual
		bestOffset := nominalOffset
		var minResidual float64 = 1e18
		var bestGainL, bestGainR float64

		for d := -16; d <= 16; d++ {
			off := nominalOffset + d
			if off < 0 || off+trimmedWmFrames > numSongFrames {
				continue
			}

			var crossL, crossR float64
			for i := 0; i < trimmedWmFrames; i++ {
				crossL += songL[off+i] * wmL[i]
				crossR += songR[off+i] * wmR[i]
			}
			gL := crossL / (wmEL + 1e-12)
			gR := crossR / (wmER + 1e-12)

			var res float64
			for i := 0; i < trimmedWmFrames; i++ {
				dL := songL[off+i] - gL*wmL[i]
				dR := songR[off+i] - gR*wmR[i]
				res += dL*dL + dR*dR
			}

			if res < minResidual {
				minResidual = res
				bestOffset = off
				bestGainL = gL
				bestGainR = gR
			}
		}

		if bestOffset < 0 {
			bestOffset = 0
		}
		actualLen := trimmedWmFrames
		if bestOffset+actualLen > numSongFrames {
			actualLen = numSongFrames - bestOffset
		}

		// Stage 1: Time-Domain Subtraction (Phase Cancellation)
		for i := 0; i < actualLen; i++ {
			songL[bestOffset+i] -= bestGainL * wmL[i]
			if mainChannels > 1 {
				songR[bestOffset+i] -= bestGainR * wmR[i]
			}
		}

		// Stage 2: Enhanced Spectral Subtraction on the watermark window (eradicates residual MP3 whisper)
		subSegL := make([]float64, actualLen)
		copy(subSegL, songL[bestOffset:bestOffset+actualLen])
		cleanSegL := spectralSubtraction(subSegL, wmL[:actualLen], bestGainL, 0.45, 0.04, float64(targetSampleRate))

		var cleanSegR []float64
		var subSegR []float64
		if mainChannels > 1 {
			subSegR = make([]float64, actualLen)
			copy(subSegR, songR[bestOffset:bestOffset+actualLen])
			cleanSegR = spectralSubtraction(subSegR, wmR[:actualLen], bestGainR, 0.45, 0.04, float64(targetSampleRate))
		}

		// Smooth boundaries with 64-sample cosine cross-fade
		for i := 0; i < actualLen; i++ {
			fade := 1.0
			if i < fadeLen {
				fade = 0.5 * (1.0 - math.Cos(math.Pi*float64(i)/float64(fadeLen)))
			} else if i > actualLen-fadeLen {
				fade = 0.5 * (1.0 - math.Cos(math.Pi*float64(actualLen-i)/float64(fadeLen)))
			}

			songL[bestOffset+i] = subSegL[i]*(1.0-fade) + cleanSegL[i]*fade
			if mainChannels > 1 {
				songR[bestOffset+i] = subSegR[i]*(1.0-fade) + cleanSegR[i]*fade
			}
		}

		progress := 65.0 + float64(idx+1)/float64(len(matches))*15.0
		sendTaskLog(task, progress, "processing", fmt.Sprintf("Processed watermark %d/%d (at %.1fs)", idx+1, len(matches), float64(bestOffset)/float64(targetSampleRate)), fmt.Sprintf("Watermark #%d cancelled at %.2fs: phase alignment aligned, spectral subtraction applied", idx+1, float64(bestOffset)/float64(targetSampleRate)))
	}

	// 6. Interleave clean channels back into float32 slice
	cleanAudio := make([]float32, len(mainAudio))
	for i := 0; i < numSongFrames; i++ {
		valL := songL[i]
		if valL > 1.0 {
			valL = 1.0
		} else if valL < -1.0 {
			valL = -1.0
		}
		cleanAudio[i*mainChannels] = float32(valL)

		if mainChannels > 1 {
			valR := songR[i]
			if valR > 1.0 {
				valR = 1.0
			} else if valR < -1.0 {
				valR = -1.0
			}
			cleanAudio[i*mainChannels+1] = float32(valR)
		}
	}

	// 7. Encode clean audio via FFmpeg pipe directly into output file
	sendTaskLog(task, 82, "processing", "Encoding clean audio...", fmt.Sprintf("Streaming uncompressed float32 samples to %s...", filepath.Base(outputFile)))
	err = encodeFloat32Audio(cleanAudio, targetSampleRate, mainChannels, outputFile, func(p float64) {
		sendTaskLog(task, 82+p*0.17, "processing", "Encoding clean audio...", fmt.Sprintf("FFmpeg audio encoding progress: %.1f%%", p))
	})
	if err != nil {
		return fmt.Errorf("failed to encode output: %v", err)
	}

	completeTask(task, outputFile)
	return nil
}

// spectralSubtraction reduces leftover watermark frequency residue via STFT.
func spectralSubtraction(signal, wmRef []float64, gain, beta, gamma, sampleRate float64) []float64 {
	const winSize = 1024
	const hopSize = 256

	n := len(signal)
	if n < winSize {
		return signal
	}

	window := make([]float64, winSize)
	for i := 0; i < winSize; i++ {
		window[i] = 0.5 * (1.0 - math.Cos(2.0*math.Pi*float64(i)/float64(winSize)))
	}

	output := make([]float64, n)
	norm := make([]float64, n)

	wmPadded := make([]float64, n)
	copy(wmPadded, wmRef)

	frameSig := make([]complex128, winSize)
	frameWM := make([]complex128, winSize)

	binHz := sampleRate / float64(winSize)

	for pos := 0; pos+winSize <= n; pos += hopSize {
		for i := 0; i < winSize; i++ {
			frameSig[i] = complex(signal[pos+i]*window[i], 0)
			frameWM[i] = complex(wmPadded[pos+i]*window[i], 0)
		}

		fftRadix2(frameSig, false)
		fftRadix2(frameWM, false)

		for i := 0; i < winSize; i++ {
			freq := float64(i) * binHz
			if i > winSize/2 {
				freq = float64(winSize-i) * binHz
			}

			magY := cmplx.Abs(frameSig[i])
			phaseY := cmplx.Phase(frameSig[i])
			magW := cmplx.Abs(frameWM[i])

			// Vocal frequency weighting:
			// Full suppression in speech band 200 Hz - 7500 Hz
			// Smooth rolloff outside speech band to preserve low sub-bass and high cymbals
			freqWeight := 1.0
			if freq < 200 {
				freqWeight = math.Max(0.1, freq/200.0)
			} else if freq > 7500 {
				freqWeight = math.Max(0.2, 1.0-(freq-7500.0)/8000.0)
			}

			sub := beta * gain * magW * freqWeight
			cleanMag := magY - sub
			floor := gamma * magY
			if cleanMag < floor {
				cleanMag = floor
			}

			frameSig[i] = cmplx.Rect(cleanMag, phaseY)
		}

		fftRadix2(frameSig, true)

		for i := 0; i < winSize; i++ {
			output[pos+i] += real(frameSig[i]) * window[i]
			norm[pos+i] += window[i] * window[i]
		}
	}

	for i := 0; i < n; i++ {
		if norm[i] > 1e-4 {
			output[i] /= norm[i]
		} else {
			output[i] = signal[i]
		}
	}

	return output
}

// detectAllWatermarkPeaks finds all watermark occurrences in song using FFT cross-correlation.
func detectAllWatermarkPeaks(song, watermark []float64, sampleRate int) []WatermarkMatch {
	N := len(song)
	M := len(watermark)
	if M == 0 || N < M {
		return nil
	}

	// For very long audio (> 5 minutes), process in overlapping segments
	const maxDirectLength = 60 * 44100
	if N > maxDirectLength*2 {
		return detectAllWatermarkPeaksSegmented(song, watermark, maxDirectLength)
	}

	matches := detectAllWatermarkPeaksSingle(song, watermark, 0, 0.15)
	if len(matches) == 0 {
		// Try slightly lower threshold in case watermark is quiet
		matches = detectAllWatermarkPeaksSingle(song, watermark, 0, 0.11)
	}
	return matches
}

func detectAllWatermarkPeaksSingle(song, watermark []float64, baseOffset int, threshold float64) []WatermarkMatch {
	N := len(song)
	M := len(watermark)
	if M == 0 || N < M {
		return nil
	}

	L := nextPowerOf2(N + M - 1)

	X := make([]complex128, L)
	for i, s := range song {
		X[i] = complex(s, 0)
	}
	W := make([]complex128, L)
	for i, s := range watermark {
		W[i] = complex(s, 0)
	}

	// FFT
	fftRadix2(X, false)
	fftRadix2(W, false)

	// Cross-correlation in frequency domain: X * conj(W)
	for i := range X {
		X[i] *= cmplx.Conj(W[i])
	}

	// IFFT
	fftRadix2(X, true)

	// Cumulative prefix energy of song
	prefE := make([]float64, N+1)
	for i, s := range song {
		prefE[i+1] = prefE[i] + s*s
	}

	var wmEnergy float64
	for _, s := range watermark {
		wmEnergy += s * s
	}
	if wmEnergy <= 1e-12 {
		return nil
	}

	var candidates []peakCandidate

	maxTau := N - M
	for tau := 1; tau < maxTau; tau++ {
		corr := real(X[tau])
		// Strict local maximum condition on the correlation curve
		if corr > real(X[tau-1]) && corr > real(X[tau+1]) && corr > 0 {
			lE := prefE[tau+M] - prefE[tau]
			denom := math.Sqrt(wmEnergy*lE + 1e-12)
			rho := corr / denom
			if rho >= threshold {
				candidates = append(candidates, peakCandidate{
					offset: tau,
					corr:   corr,
					rho:    rho,
				})
			}
		}
	}

	// Sort candidates by correlation descending
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].rho > candidates[j].rho
	})

	// Greedy selection: ensure non-overlapping watermark occurrences (at least M/2 apart)
	var matches []WatermarkMatch
	var selectedOffsets []int

	for _, c := range candidates {
		overlap := false
		for _, prev := range selectedOffsets {
			diff := c.offset - prev
			if diff < 0 {
				diff = -diff
			}
			if diff < M/2 {
				overlap = true
				break
			}
		}

		if !overlap {
			selectedOffsets = append(selectedOffsets, c.offset)

			var crossSum float64
			for i := 0; i < M; i++ {
				crossSum += song[c.offset+i] * watermark[i]
			}
			gain := crossSum / (wmEnergy + 1e-12)

			matches = append(matches, WatermarkMatch{
				Offset:      baseOffset + c.offset,
				Correlation: c.rho,
				Gain:        gain,
			})
		}
	}

	// Re-sort matches chronologically by offset
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Offset < matches[j].Offset
	})

	return matches
}

func detectAllWatermarkPeaksSegmented(song, watermark []float64, segSize int) []WatermarkMatch {
	N := len(song)
	M := len(watermark)
	step := segSize - M

	var allMatches []WatermarkMatch
	usedOffsets := make(map[int]bool)

	for start := 0; start < N; start += step {
		end := start + segSize
		if end > N {
			end = N
		}
		if end-start < M {
			break
		}

		subSong := song[start:end]
		matches := detectAllWatermarkPeaksSingle(subSong, watermark, start, 0.15)
		if len(matches) == 0 {
			matches = detectAllWatermarkPeaksSingle(subSong, watermark, start, 0.11)
		}

		for _, m := range matches {
			tooClose := false
			for off := range usedOffsets {
				if intAbs(m.Offset-off) < M/2 {
					tooClose = true
					break
				}
			}
			if !tooClose {
				usedOffsets[m.Offset] = true
				allMatches = append(allMatches, m)
			}
		}
	}

	sort.Slice(allMatches, func(i, j int) bool {
		return allMatches[i].Offset < allMatches[j].Offset
	})

	return allMatches
}

// decodeAudioFile decodes any audio file into float32 samples at the given sample rate.
// It supports standard media files (MP3, WAV, FLAC, AAC, OGG, etc.) and custom SRM format.
func decodeAudioFile(path string, sampleRate int, forceMono bool) ([]float32, int, error) {
	// First check if the file is an SRM binary file
	srmData, err := parseSRMFile(path)
	if err == nil {
		if forceMono || srmData.Channels == 1 {
			mono := make([]float32, srmData.NumSamples)
			for i := 0; i < srmData.NumSamples; i++ {
				mono[i] = 0.5 * (srmData.Ch0[i] + srmData.Ch1[i])
			}
			return mono, 1, nil
		}
		// Interleave stereo [L0, R0, L1, R1, ...]
		interleaved := make([]float32, srmData.NumSamples*2)
		for i := 0; i < srmData.NumSamples; i++ {
			interleaved[i*2] = srmData.Ch0[i]
			interleaved[i*2+1] = srmData.Ch1[i]
		}
		return interleaved, 2, nil
	}

	ffmpegPath := findFFmpeg()

	channels := 2
	if forceMono {
		channels = 1
	} else {
		info, err := ProbeFile(path)
		if err == nil {
			for _, st := range info.Streams {
				if st.Type == "audio" && st.Channels == 1 {
					channels = 1
					break
				}
			}
		}
	}

	// Standard decoding via FFmpeg pipe
	args := []string{
		"-v", "error",
		"-i", path,
		"-f", "f32le",
		"-acodec", "pcm_f32le",
		"-ar", strconv.Itoa(sampleRate),
		"-ac", strconv.Itoa(channels),
		"pipe:1",
	}

	cmd := exec.Command(ffmpegPath, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err == nil && stdout.Len() > 0 {
		return bytesToFloat32(stdout.Bytes()), channels, nil
	}

	// Fallback 1: Raw 16-bit PCM (.s16le, .raw, .pcm)
	argsRawS16 := []string{
		"-v", "error",
		"-f", "s16le",
		"-ar", strconv.Itoa(sampleRate),
		"-ac", strconv.Itoa(channels),
		"-i", path,
		"-f", "f32le",
		"-acodec", "pcm_f32le",
		"-ar", strconv.Itoa(sampleRate),
		"-ac", strconv.Itoa(channels),
		"pipe:1",
	}
	cmd = exec.Command(ffmpegPath, argsRawS16...)
	stdout.Reset()
	if err := cmd.Run(); err == nil && stdout.Len() > 0 {
		return bytesToFloat32(stdout.Bytes()), channels, nil
	}

	// Fallback 2: Raw 32-bit float PCM
	argsRawF32 := []string{
		"-v", "error",
		"-f", "f32le",
		"-ar", strconv.Itoa(sampleRate),
		"-ac", strconv.Itoa(channels),
		"-i", path,
		"-f", "f32le",
		"-acodec", "pcm_f32le",
		"-ar", strconv.Itoa(sampleRate),
		"-ac", strconv.Itoa(channels),
		"pipe:1",
	}
	cmd = exec.Command(ffmpegPath, argsRawF32...)
	stdout.Reset()
	if err := cmd.Run(); err == nil && stdout.Len() > 0 {
		return bytesToFloat32(stdout.Bytes()), channels, nil
	}

	errStr := strings.TrimSpace(stderr.String())
	if errStr == "" {
		errStr = "unrecognized audio format"
	}
	return nil, channels, fmt.Errorf("%s", errStr)
}

func bytesToFloat32(data []byte) []float32 {
	num := len(data) / 4
	samples := make([]float32, num)
	for i := 0; i < num; i++ {
		bits := binary.LittleEndian.Uint32(data[i*4 : (i+1)*4])
		samples[i] = math.Float32frombits(bits)
	}
	return samples
}

func float32ToBytes(samples []float32) []byte {
	buf := make([]byte, len(samples)*4)
	for i, s := range samples {
		bits := math.Float32bits(s)
		binary.LittleEndian.PutUint32(buf[i*4:(i+1)*4], bits)
	}
	return buf
}

func encodeFloat32Audio(samples []float32, sampleRate, channels int, outputFile string, onProgress func(float64)) error {
	ffmpegPath := findFFmpeg()
	ext := strings.ToLower(filepath.Ext(outputFile))

	var codecArgs []string
	switch ext {
	case ".mp3":
		codecArgs = []string{"-acodec", "libmp3lame", "-b:a", "320k"}
	case ".wav":
		codecArgs = []string{"-acodec", "pcm_s16le"}
	case ".flac":
		codecArgs = []string{"-acodec", "flac"}
	case ".aac", ".m4a":
		codecArgs = []string{"-acodec", "aac", "-b:a", "256k"}
	case ".ogg":
		codecArgs = []string{"-acodec", "libvorbis", "-q:a", "6"}
	default:
		codecArgs = []string{"-acodec", "libmp3lame", "-b:a", "320k"}
	}

	args := []string{
		"-v", "error",
		"-f", "f32le",
		"-ar", strconv.Itoa(sampleRate),
		"-ac", strconv.Itoa(channels),
		"-i", "pipe:0",
	}
	args = append(args, codecArgs...)
	args = append(args, "-y", outputFile)

	cmd := exec.Command(ffmpegPath, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to open ffmpeg stdin: %v", err)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg encoder: %v", err)
	}

	rawBytes := float32ToBytes(samples)
	chunkSize := 65536
	totalBytes := len(rawBytes)
	written := 0

	for written < totalBytes {
		end := written + chunkSize
		if end > totalBytes {
			end = totalBytes
		}
		n, err := stdin.Write(rawBytes[written:end])
		if err != nil && err != io.EOF {
			break
		}
		written += n
		if onProgress != nil && totalBytes > 0 {
			onProgress(float64(written) / float64(totalBytes) * 100)
		}
	}
	_ = stdin.Close()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg encoding error: %v - %s", err, stderr.String())
	}
	return nil
}

func fftRadix2(a []complex128, inverse bool) {
	n := len(a)
	if n <= 1 {
		return
	}

	j := 0
	for i := 0; i < n; i++ {
		if i < j {
			a[i], a[j] = a[j], a[i]
		}
		bit := n >> 1
		for ; j&bit != 0; bit >>= 1 {
			j ^= bit
		}
		j ^= bit
	}

	sign := -1.0
	if inverse {
		sign = 1.0
	}

	for length := 2; length <= n; length <<= 1 {
		half := length >> 1
		angle := sign * 2.0 * math.Pi / float64(length)
		wStep := cmplx.Rect(1, angle)
		for i := 0; i < n; i += length {
			w := complex(1, 0)
			for k := 0; k < half; k++ {
				u := a[i+k]
				v := a[i+k+half] * w
				a[i+k] = u + v
				a[i+k+half] = u - v
				w *= wStep
			}
		}
	}

	if inverse {
		invN := complex(1.0/float64(n), 0)
		for i := range a {
			a[i] *= invN
		}
	}
}

func nextPowerOf2(n int) int {
	p := 1
	for p < n {
		p <<= 1
	}
	return p
}

func intAbs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
