package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// browseEntry represents a file or directory entry in the browse response.
type browseEntry struct {
	Name     string `json:"name"`
	IsDir    bool   `json:"isDir"`
	Size     int64  `json:"size"`
	Ext      string `json:"ext"`
	Modified string `json:"modified"`
}

// QuickAccess represents a quick navigation folder.
type QuickAccess struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// cleanTempIfUpload cleans up temporary uploaded files created by /api/upload.
func cleanTempIfUpload(path string) {
	if path == "" {
		return
	}
	tempDir := filepath.Clean(os.TempDir())
	cleanP := filepath.Clean(path)
	if strings.HasPrefix(cleanP, tempDir) && strings.Contains(filepath.Base(cleanP), "edeth-upload-") {
		_ = os.Remove(cleanP)
	}
}

// handleUpload streams large multipart file uploads directly to temporary files with O(1) RAM.
func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Bounded streaming: use MultipartReader to avoid loading entire file into RAM
	mr, err := r.MultipartReader()
	if err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid multipart form: %v", err))
		return
	}

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			s.writeError(w, http.StatusBadRequest, fmt.Sprintf("failed to read multipart stream: %v", err))
			return
		}

		filename := part.FileName()
		if filename == "" {
			_ = part.Close()
			continue
		}

		ext := filepath.Ext(filename)
		tempFile, err := os.CreateTemp("", "edeth-upload-*"+ext)
		if err != nil {
			_ = part.Close()
			s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create temp file: %v", err))
			return
		}

		// Stream directly from HTTP body to disk using a 32KB buffer
		buf := make([]byte, 32*1024)
		n, copyErr := io.CopyBuffer(tempFile, part, buf)
		_ = tempFile.Close()
		_ = part.Close()

		if copyErr != nil {
			_ = os.Remove(tempFile.Name())
			s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to write upload stream: %v", copyErr))
			return
		}

		s.writeJSON(w, http.StatusOK, map[string]interface{}{
			"filePath": tempFile.Name(),
			"fileName": filename,
			"size":     n,
		})
		return
	}

	s.writeError(w, http.StatusBadRequest, "no file field found in request")
}

// handleBrowse handles GET /api/browse?path=... cross-platform.
func (s *Server) handleBrowse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	dirPath := cleanPath(r.URL.Query().Get("path"))
	if dirPath == "" {
		if cwd, err := os.Getwd(); err == nil {
			dirPath = cwd
		} else if home, err := os.UserHomeDir(); err == nil {
			dirPath = home
		} else {
			dirPath = "/"
		}
	}

	// Windows drive root fix: "C:" -> "C:\"
	if runtime.GOOS == "windows" && len(dirPath) == 2 && dirPath[1] == ':' {
		dirPath = dirPath + "\\"
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("failed to read directory: %v", err))
		return
	}

	var result []browseEntry
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		result = append(result, browseEntry{
			Name:     entry.Name(),
			IsDir:    entry.IsDir(),
			Size:     info.Size(),
			Ext:      filepath.Ext(entry.Name()),
			Modified: info.ModTime().Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].IsDir != result[j].IsDir {
			return result[i].IsDir
		}
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})

	parent := filepath.Dir(dirPath)
	if parent == dirPath {
		if runtime.GOOS == "windows" && len(dirPath) == 3 && dirPath[1] == ':' {
			parent = ""
		} else if runtime.GOOS != "windows" && dirPath == "/" {
			parent = ""
		}
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"path":    dirPath,
		"parent":  parent,
		"entries": result,
	})
}

// handleDrives handles GET /api/drives cross-platform (Windows drives or Linux/Docker mount points).
func (s *Server) handleDrives(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var drives []string
	if runtime.GOOS == "windows" {
		for c := 'A'; c <= 'Z'; c++ {
			drivePath := fmt.Sprintf("%c:\\", c)
			if _, err := os.Stat(drivePath); err == nil {
				drives = append(drives, fmt.Sprintf("%c:", c))
			}
		}
	} else {
		// Linux, macOS, and Docker environments
		candidates := []string{"/", "/workspace", "/data", "/app", "/tmp"}
		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				drives = append(drives, candidate)
			}
		}
		if len(drives) == 0 {
			drives = []string{"/"}
		}
	}

	home, _ := os.UserHomeDir()
	var quick []QuickAccess

	// Add current working directory as first quick access
	if cwd, err := os.Getwd(); err == nil {
		quick = append(quick, QuickAccess{Name: "Workspace (CWD)", Path: cwd})
	}

	if home != "" {
		for _, name := range []string{"Desktop", "Downloads", "Documents", "Music", "Videos"} {
			p := filepath.Join(home, name)
			if _, err := os.Stat(p); err == nil {
				quick = append(quick, QuickAccess{Name: name, Path: p})
			}
		}
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"drives": drives,
		"home":   home,
		"quick":  quick,
	})
}

// handleMediaInfo handles GET /api/media-info?path=...
func (s *Server) handleMediaInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	filePath := cleanPath(r.URL.Query().Get("path"))
	if filePath == "" {
		s.writeError(w, http.StatusBadRequest, "missing path parameter")
		return
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		s.writeError(w, http.StatusNotFound, "file not found")
		return
	}

	info, err := ProbeFile(filePath)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to probe file: %v", err))
		return
	}

	s.writeJSON(w, http.StatusOK, info)
}

// watermarkRemoveRequest is the request body for watermark removal.
type watermarkRemoveRequest struct {
	InputFile     string `json:"inputFile"`
	WatermarkFile string `json:"watermarkFile"`
	OutputFile    string `json:"outputFile"`
}

// handleWatermarkRemove handles POST /api/watermark/remove
func (s *Server) handleWatermarkRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req watermarkRemoveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.InputFile = cleanPath(req.InputFile)
	req.WatermarkFile = cleanPath(req.WatermarkFile)
	req.OutputFile = cleanPath(req.OutputFile)

	if req.InputFile == "" || req.WatermarkFile == "" {
		s.writeError(w, http.StatusBadRequest, "inputFile and watermarkFile are required")
		return
	}

	if _, err := os.Stat(req.InputFile); os.IsNotExist(err) {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("input file does not exist: %s", req.InputFile))
		return
	}
	if _, err := os.Stat(req.WatermarkFile); os.IsNotExist(err) {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("watermark file does not exist: %s", req.WatermarkFile))
		return
	}

	if req.OutputFile == "" {
		ext := filepath.Ext(req.InputFile)
		base := strings.TrimSuffix(req.InputFile, ext)
		req.OutputFile = base + "_clean" + ext
	}

	task := s.createTask()

	go func() {
		defer cleanTempIfUpload(req.InputFile)
		defer cleanTempIfUpload(req.WatermarkFile)

		err := RemoveWatermark(req.InputFile, req.WatermarkFile, req.OutputFile, task)
		if err != nil {
			failTask(task, fmt.Sprintf("watermark removal failed: %v", err))
		}
	}()

	s.writeJSON(w, http.StatusOK, map[string]string{
		"taskId":     task.ID,
		"outputFile": req.OutputFile,
	})
}

// subtitleAddRequest is the request body for subtitle addition.
type subtitleAddRequest struct {
	VideoFile    string `json:"videoFile"`
	SubtitleFile string `json:"subtitleFile"`
	OutputFile   string `json:"outputFile"`
	BurnIn       bool   `json:"burnIn"`
}

// handleSubtitleAdd handles POST /api/subtitle/add
func (s *Server) handleSubtitleAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req subtitleAddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.VideoFile = cleanPath(req.VideoFile)
	req.SubtitleFile = cleanPath(req.SubtitleFile)
	req.OutputFile = cleanPath(req.OutputFile)

	if req.VideoFile == "" || req.SubtitleFile == "" {
		s.writeError(w, http.StatusBadRequest, "videoFile and subtitleFile are required")
		return
	}

	if _, err := os.Stat(req.VideoFile); os.IsNotExist(err) {
		s.writeError(w, http.StatusBadRequest, "video file does not exist")
		return
	}
	if _, err := os.Stat(req.SubtitleFile); os.IsNotExist(err) {
		s.writeError(w, http.StatusBadRequest, "subtitle file does not exist")
		return
	}

	if req.OutputFile == "" {
		ext := filepath.Ext(req.VideoFile)
		base := strings.TrimSuffix(req.VideoFile, ext)
		req.OutputFile = base + "_subtitled" + ext
	}

	task := s.createTask()

	go func() {
		defer cleanTempIfUpload(req.VideoFile)
		defer cleanTempIfUpload(req.SubtitleFile)

		info, err := ProbeFile(req.VideoFile)
		duration := 0.0
		if err == nil {
			duration = info.Duration
		}

		sendTaskLog(task, 5, "processing", "Adding subtitles...", "Probing media streams and preparing FFmpeg filtergraph...")

		var args []string
		if req.BurnIn {
			escapedSubPath := EscapeFilterPath(req.SubtitleFile)
			args = []string{
				"-v", "error",
				"-i", req.VideoFile,
				"-vf", fmt.Sprintf("subtitles='%s'", escapedSubPath),
				"-c:v", "libx264",
				"-preset", "fast",
				"-c:a", "copy",
				"-y", req.OutputFile,
			}
		} else {
			outExt := strings.ToLower(filepath.Ext(req.OutputFile))
			subCodec := "mov_text"
			if outExt == ".mkv" {
				subCodec = "srt"
			} else if outExt == ".webm" {
				subCodec = "webvtt"
			}

			args = []string{
				"-v", "error",
				"-i", req.VideoFile,
				"-i", req.SubtitleFile,
				"-c:v", "copy",
				"-c:a", "copy",
				"-c:s", subCodec,
				"-y", req.OutputFile,
			}
		}

		err = RunFFmpegContext(task.ctx, args, duration, func(progress float64, line string) {
			sendTaskLog(task, 5+progress*0.9, "processing", "Adding subtitles...", line)
		})
		if err != nil {
			failTask(task, fmt.Sprintf("subtitle add failed: %v", err))
			return
		}

		completeTask(task, req.OutputFile)
	}()

	s.writeJSON(w, http.StatusOK, map[string]string{
		"taskId":     task.ID,
		"outputFile": req.OutputFile,
	})
}
// trimRequest is the request body for trimming.
type trimRequest struct {
	InputFile  string `json:"inputFile"`
	OutputFile string `json:"outputFile"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
}

// handleTrim handles POST /api/trim
func (s *Server) handleTrim(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req trimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.InputFile = cleanPath(req.InputFile)
	req.OutputFile = cleanPath(req.OutputFile)
	req.StartTime = strings.TrimSpace(req.StartTime)
	req.EndTime = strings.TrimSpace(req.EndTime)

	if req.InputFile == "" || req.StartTime == "" || req.EndTime == "" {
		s.writeError(w, http.StatusBadRequest, "inputFile, startTime, and endTime are required")
		return
	}

	if _, err := os.Stat(req.InputFile); os.IsNotExist(err) {
		s.writeError(w, http.StatusBadRequest, "input file does not exist")
		return
	}

	startSec := parseTimeToSeconds(req.StartTime)
	endSec := parseTimeToSeconds(req.EndTime)
	duration := endSec - startSec
	if duration <= 0 {
		s.writeError(w, http.StatusBadRequest, "end time must be greater than start time")
		return
	}

	if req.OutputFile == "" {
		ext := filepath.Ext(req.InputFile)
		base := strings.TrimSuffix(req.InputFile, ext)
		req.OutputFile = base + "_trimmed" + ext
	}

	task := s.createTask()

	go func() {
		defer cleanTempIfUpload(req.InputFile)

		sendTaskLog(task, 5, "processing", "Trimming media...", fmt.Sprintf("Cutting from %.3fs for duration %.3fs", startSec, duration))

		args := []string{
			"-v", "error",
			"-ss", fmt.Sprintf("%.3f", startSec),
			"-i", req.InputFile,
			"-t", fmt.Sprintf("%.3f", duration),
			"-c", "copy",
			"-avoid_negative_ts", "make_zero",
			"-y", req.OutputFile,
		}

		err := RunFFmpegContext(task.ctx, args, duration, func(progress float64, line string) {
			sendTaskLog(task, 5+progress*0.9, "processing", "Trimming media...", line)
		})
		if err != nil {
			failTask(task, fmt.Sprintf("trim failed: %v", err))
			return
		}

		completeTask(task, req.OutputFile)
	}()

	s.writeJSON(w, http.StatusOK, map[string]string{
		"taskId":     task.ID,
		"outputFile": req.OutputFile,
	})
}
// audioExtractRequest is the request body for audio extraction.
type audioExtractRequest struct {
	VideoFile  string `json:"videoFile"`
	OutputFile string `json:"outputFile"`
	Format     string `json:"format"`
}

// handleAudioExtract handles POST /api/audio/extract
func (s *Server) handleAudioExtract(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req audioExtractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.VideoFile = cleanPath(req.VideoFile)
	req.OutputFile = cleanPath(req.OutputFile)

	if req.VideoFile == "" {
		s.writeError(w, http.StatusBadRequest, "videoFile is required")
		return
	}

	if _, err := os.Stat(req.VideoFile); os.IsNotExist(err) {
		s.writeError(w, http.StatusBadRequest, "video file does not exist")
		return
	}

	if req.Format == "" {
		req.Format = "mp3"
	}

	var codec string
	switch strings.ToLower(req.Format) {
	case "mp3":
		codec = "libmp3lame"
	case "wav":
		codec = "pcm_s16le"
	case "aac":
		codec = "aac"
	case "flac":
		codec = "flac"
	case "ogg":
		codec = "libvorbis"
	default:
		codec = "libmp3lame"
		req.Format = "mp3"
	}

	if req.OutputFile == "" {
		base := strings.TrimSuffix(req.VideoFile, filepath.Ext(req.VideoFile))
		req.OutputFile = base + "." + req.Format
	}

	task := s.createTask()

	go func() {
		defer cleanTempIfUpload(req.VideoFile)

		info, err := ProbeFile(req.VideoFile)
		duration := 0.0
		if err == nil {
			duration = info.Duration
		}

		sendTaskLog(task, 5, "processing", "Extracting audio...", fmt.Sprintf("Extracting audio stream using codec %s", codec))

		args := []string{
			"-v", "error",
			"-i", req.VideoFile,
			"-vn",
			"-map", "0:a:0?",
			"-acodec", codec,
			"-y", req.OutputFile,
		}

		err = RunFFmpegContext(task.ctx, args, duration, func(progress float64, line string) {
			sendTaskLog(task, 5+progress*0.9, "processing", "Extracting audio...", line)
		})
		if err != nil {
			failTask(task, fmt.Sprintf("audio extraction failed: %v", err))
			return
		}

		completeTask(task, req.OutputFile)
	}()

	s.writeJSON(w, http.StatusOK, map[string]string{
		"taskId":     task.ID,
		"outputFile": req.OutputFile,
	})
}
// audioMergeRequest is the request body for merging audio into video.
type audioMergeRequest struct {
	VideoFile  string  `json:"videoFile"`
	AudioFile  string  `json:"audioFile"`
	OutputFile string  `json:"outputFile"`
	Loop       bool    `json:"loop"`
	StartAt    float64 `json:"startAt"`
}

// handleAudioMerge handles POST /api/audio/merge
func (s *Server) handleAudioMerge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req audioMergeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.VideoFile = cleanPath(req.VideoFile)
	req.AudioFile = cleanPath(req.AudioFile)
	req.OutputFile = cleanPath(req.OutputFile)

	if req.VideoFile == "" || req.AudioFile == "" {
		s.writeError(w, http.StatusBadRequest, "videoFile and audioFile are required")
		return
	}

	if _, err := os.Stat(req.VideoFile); os.IsNotExist(err) {
		s.writeError(w, http.StatusBadRequest, "video file does not exist")
		return
	}
	if _, err := os.Stat(req.AudioFile); os.IsNotExist(err) {
		s.writeError(w, http.StatusBadRequest, "audio file does not exist")
		return
	}

	if req.OutputFile == "" {
		ext := filepath.Ext(req.VideoFile)
		base := strings.TrimSuffix(req.VideoFile, ext)
		req.OutputFile = base + "_merged" + ext
	}

	task := s.createTask()

	go func() {
		defer cleanTempIfUpload(req.VideoFile)
		defer cleanTempIfUpload(req.AudioFile)

		videoInfo, err := ProbeFile(req.VideoFile)
		duration := 0.0
		hasAudio := false
		if err == nil {
			duration = videoInfo.Duration
			for _, stream := range videoInfo.Streams {
				if stream.Type == "audio" {
					hasAudio = true
					break
				}
			}
		}

		sendTaskLog(task, 5, "processing", "Merging audio...", "Configuring FFmpeg audio mixing pipeline...")

		var args []string
		if !hasAudio {
			if req.Loop {
				args = []string{
					"-v", "error",
					"-i", req.VideoFile,
					"-stream_loop", "-1",
					"-i", req.AudioFile,
					"-c:v", "copy",
					"-c:a", "aac",
					"-map", "0:v:0",
					"-map", "1:a:0",
					"-shortest",
					"-y", req.OutputFile,
				}
			} else if req.StartAt > 0 {
				delayMs := int(req.StartAt * 1000)
				args = []string{
					"-v", "error",
					"-i", req.VideoFile,
					"-i", req.AudioFile,
					"-filter_complex", fmt.Sprintf("[1:a]adelay=%d:all=1[a1]", delayMs),
					"-c:v", "copy",
					"-c:a", "aac",
					"-map", "0:v:0",
					"-map", "[a1]",
					"-shortest",
					"-y", req.OutputFile,
				}
			} else {
				args = []string{
					"-v", "error",
					"-i", req.VideoFile,
					"-i", req.AudioFile,
					"-c:v", "copy",
					"-c:a", "aac",
					"-map", "0:v:0",
					"-map", "1:a:0",
					"-shortest",
					"-y", req.OutputFile,
				}
			}
		} else {
			if req.Loop {
				args = []string{
					"-v", "error",
					"-i", req.VideoFile,
					"-stream_loop", "-1",
					"-i", req.AudioFile,
					"-filter_complex",
					fmt.Sprintf("[1:a]atrim=0:%.3f[a1];[0:a][a1]amix=inputs=2:duration=first", duration),
					"-c:v", "copy",
					"-c:a", "aac",
					"-y", req.OutputFile,
				}
			} else if req.StartAt > 0 {
				delayMs := int(req.StartAt * 1000)
				args = []string{
					"-v", "error",
					"-i", req.VideoFile,
					"-i", req.AudioFile,
					"-filter_complex",
					fmt.Sprintf("[1:a]adelay=%d:all=1[a1];[0:a][a1]amix=inputs=2:duration=first", delayMs),
					"-c:v", "copy",
					"-c:a", "aac",
					"-y", req.OutputFile,
				}
			} else {
				args = []string{
					"-v", "error",
					"-i", req.VideoFile,
					"-i", req.AudioFile,
					"-filter_complex",
					"[0:a][1:a]amix=inputs=2:duration=first",
					"-c:v", "copy",
					"-c:a", "aac",
					"-y", req.OutputFile,
				}
			}
		}

		err = RunFFmpegContext(task.ctx, args, duration, func(progress float64, line string) {
			sendTaskLog(task, 5+progress*0.9, "processing", "Merging audio...", line)
		})
		if err != nil {
			failTask(task, fmt.Sprintf("audio merge failed: %v", err))
			return
		}

		completeTask(task, req.OutputFile)
	}()

	s.writeJSON(w, http.StatusOK, map[string]string{
		"taskId":     task.ID,
		"outputFile": req.OutputFile,
	})
}

// srmToWAV converts parsed SRM planar float32 audio to a standard 16-bit PCM WAV byte buffer.
func srmToWAV(srm *SRMData) []byte {
	channels := srm.Channels
	if channels < 1 {
		channels = 1
	}
	sampleRate := srm.SampleRate
	if sampleRate <= 0 {
		sampleRate = 44100
	}
	numSamples := srm.NumSamples
	if numSamples <= 0 {
		numSamples = len(srm.Ch0)
	}

	dataSize := numSamples * channels * 2
	buf := make([]byte, 44+dataSize)

	// RIFF header
	copy(buf[0:4], "RIFF")
	binary.LittleEndian.PutUint32(buf[4:8], uint32(36+dataSize))
	copy(buf[8:12], "WAVE")

	// fmt subchunk
	copy(buf[12:16], "fmt ")
	binary.LittleEndian.PutUint32(buf[16:20], 16) // Subchunk1Size (16 for PCM)
	binary.LittleEndian.PutUint16(buf[20:22], 1)  // AudioFormat (1 = PCM)
	binary.LittleEndian.PutUint16(buf[22:24], uint16(channels))
	binary.LittleEndian.PutUint32(buf[24:28], uint32(sampleRate))
	byteRate := uint32(sampleRate * channels * 2)
	binary.LittleEndian.PutUint32(buf[28:32], byteRate)
	blockAlign := uint16(channels * 2)
	binary.LittleEndian.PutUint16(buf[32:34], blockAlign)
	binary.LittleEndian.PutUint16(buf[34:36], 16) // BitsPerSample

	// data subchunk
	copy(buf[36:40], "data")
	binary.LittleEndian.PutUint32(buf[40:44], uint32(dataSize))

	// PCM samples
	offset := 44
	for i := 0; i < numSamples; i++ {
		var v0 float32
		if i < len(srm.Ch0) {
			v0 = srm.Ch0[i]
		}
		if v0 > 1.0 {
			v0 = 1.0
		} else if v0 < -1.0 {
			v0 = -1.0
		}
		s0 := int16(v0 * 32767.0)
		binary.LittleEndian.PutUint16(buf[offset:offset+2], uint16(s0))
		offset += 2

		if channels > 1 {
			var v1 float32
			if i < len(srm.Ch1) {
				v1 = srm.Ch1[i]
			} else {
				v1 = v0
			}
			if v1 > 1.0 {
				v1 = 1.0
			} else if v1 < -1.0 {
				v1 = -1.0
			}
			s1 := int16(v1 * 32767.0)
			binary.LittleEndian.PutUint16(buf[offset:offset+2], uint16(s1))
			offset += 2
		}
	}

	return buf
}

// handleStream serves media files for browser playback with HTTP Range request support,
// and transcodes custom binary formats like .srm on-the-fly to WAV for seamless HTML5 <audio> streaming.
func (s *Server) handleStream(w http.ResponseWriter, r *http.Request) {
	filePath := cleanPath(r.URL.Query().Get("path"))
	if filePath == "" {
		s.writeError(w, http.StatusBadRequest, "path parameter required")
		return
	}

	info, err := os.Stat(filePath)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "file not found")
		return
	}

	if info.IsDir() {
		s.writeError(w, http.StatusBadRequest, "path is a directory")
		return
	}

	// Handle .srm file streaming on-the-fly
	if strings.ToLower(filepath.Ext(filePath)) == ".srm" {
		srm, err := parseSRMFile(filePath)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to parse SRM file: %v", err))
			return
		}

		wavData := srmToWAV(srm)
		w.Header().Set("Content-Type", "audio/wav")
		http.ServeContent(w, r, filepath.Base(filePath)+".wav", info.ModTime(), bytes.NewReader(wavData))
		return
	}

	// Serve standard media with full byte-range support (HTTP 206) for smooth seeking
	http.ServeFile(w, r, filePath)
}

// videoCompressRequest is the request payload for smart video compression.
type videoCompressRequest struct {
	VideoFile    string  `json:"videoFile"`
	OutputFile   string  `json:"outputFile"`
	Preset       string  `json:"preset"`       // "discord25", "telegram50", "web720p", "web1080p", "ultracompact", "custom"
	TargetSizeMB float64 `json:"targetSizeMB"` // Target size in MB
	Resolution   string  `json:"resolution"`   // "original", "1080p", "720p", "480p", "360p"
	Codec        string  `json:"codec"`        // "h264", "h265"
	AudioBitrate string  `json:"audioBitrate"` // "128k", "96k", "64k"
}

// handleVideoCompress handles POST /api/video/compress.
func (s *Server) handleVideoCompress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req videoCompressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.VideoFile = cleanPath(req.VideoFile)
	req.OutputFile = cleanPath(req.OutputFile)

	if req.VideoFile == "" {
		s.writeError(w, http.StatusBadRequest, "videoFile is required")
		return
	}

	if _, err := os.Stat(req.VideoFile); os.IsNotExist(err) {
		s.writeError(w, http.StatusBadRequest, "video file does not exist")
		return
	}

	if req.OutputFile == "" {
		base := strings.TrimSuffix(req.VideoFile, filepath.Ext(req.VideoFile))
		req.OutputFile = base + "_compressed.mp4"
	}

	task := s.createTask()

	go func() {
		defer cleanTempIfUpload(req.VideoFile)

		info, err := ProbeFile(req.VideoFile)
		duration := 0.0
		if err == nil {
			duration = info.Duration
		}

		sendTaskLog(task, 2, "processing", "Analyzing video stream...", fmt.Sprintf("Source duration: %.1fs", duration))

		// Target size logic
		targetMB := req.TargetSizeMB
		switch strings.ToLower(req.Preset) {
		case "discord25":
			targetMB = 24.0 // safe margin for 25MB limit
		case "telegram50":
			targetMB = 48.5 // safe margin for 50MB limit
		}

		videoCodec := "libx264"
		if strings.ToLower(req.Codec) == "h265" || strings.ToLower(req.Codec) == "hevc" {
			videoCodec = "libx265"
		}

		args := []string{"-v", "error", "-i", req.VideoFile}

		// Resolution filter
		res := strings.ToLower(req.Resolution)
		if res == "" && (req.Preset == "web720p" || (targetMB > 0 && targetMB <= 25 && duration > 120)) {
			res = "720p"
		} else if res == "" && req.Preset == "web1080p" {
			res = "1080p"
		}

		var height string
		switch res {
		case "1080p":
			height = "1080"
		case "720p":
			height = "720"
		case "480p":
			height = "480"
		case "360p":
			height = "360"
		}
		if height != "" {
			args = append(args, "-vf", fmt.Sprintf("scale=-2:%s", height))
		}

		// Audio Bitrate
		audioKbps := 128
		if req.AudioBitrate == "96k" || (req.AudioBitrate == "" && targetMB > 0 && targetMB <= 25) {
			audioKbps = 96
		} else if req.AudioBitrate == "64k" || (req.AudioBitrate == "" && targetMB > 0 && targetMB <= 10) {
			audioKbps = 64
		}

		if targetMB > 0 && duration > 0 {
			totalKbps := (targetMB * 8192) / duration
			videoKbps := totalKbps - float64(audioKbps)
			if videoKbps < 60 {
				videoKbps = 60
			}

			sendTaskLog(task, 5, "processing", "Calculating optimal bitrate...",
				fmt.Sprintf("Target: %.1f MB | Video Bitrate: %d kbps | Audio Bitrate: %d kbps", targetMB, int(videoKbps), audioKbps))

			args = append(args,
				"-c:v", videoCodec,
				"-b:v", fmt.Sprintf("%dk", int(videoKbps)),
				"-maxrate", fmt.Sprintf("%dk", int(videoKbps*1.35)),
				"-bufsize", fmt.Sprintf("%dk", int(videoKbps*2)),
				"-preset", "medium",
				"-c:a", "aac",
				"-b:a", fmt.Sprintf("%dk", audioKbps),
				"-movflags", "+faststart",
				"-y", req.OutputFile,
			)
		} else {
			crf := "23"
			switch strings.ToLower(req.Preset) {
			case "ultracompact":
				crf = "28"
			case "web720p":
				crf = "24"
			case "web1080p":
				crf = "22"
			}

			sendTaskLog(task, 5, "processing", "Compressing with CRF mode...",
				fmt.Sprintf("Preset: %s | CRF: %s | Codec: %s", req.Preset, crf, videoCodec))

			args = append(args,
				"-c:v", videoCodec,
				"-crf", crf,
				"-preset", "medium",
				"-c:a", "aac",
				"-b:a", fmt.Sprintf("%dk", audioKbps),
				"-movflags", "+faststart",
				"-y", req.OutputFile,
			)
		}

		err = RunFFmpegContext(task.ctx, args, duration, func(progress float64, line string) {
			sendTaskLog(task, 5+progress*0.9, "compressing", fmt.Sprintf("Compressing: %.1f%%", progress), line)
		})
		if err != nil {
			failTask(task, fmt.Sprintf("compression failed: %v", err))
			return
		}

		outStat, _ := os.Stat(req.OutputFile)
		msg := "Compression complete"
		if outStat != nil {
			msg = fmt.Sprintf("Compression complete (%.2f MB)", float64(outStat.Size())/(1024*1024))
		}
		sendTaskLog(task, 100, "complete", msg, msg)
		completeTask(task, req.OutputFile)
	}()

	s.writeJSON(w, http.StatusOK, map[string]string{
		"taskId":     task.ID,
		"outputFile": req.OutputFile,
	})
}
