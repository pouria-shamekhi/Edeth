package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// cleanPath strips leading/trailing quotes and whitespace, and cleans the path cross-platform.
func cleanPath(p string) string {
	p = strings.TrimSpace(p)
	p = strings.Trim(p, "\"'`")
	if p == "" {
		return ""
	}
	return filepath.Clean(p)
}

// EscapeFilterPath escapes file paths for use in FFmpeg filtergraphs cross-platform.
// On Windows, drive colons must be escaped (e.g. C\:), and single quotes must be escaped.
func EscapeFilterPath(path string) string {
	p := filepath.ToSlash(cleanPath(path))
	if runtime.GOOS == "windows" {
		p = strings.ReplaceAll(p, ":", "\\:")
	}
	p = strings.ReplaceAll(p, "'", "\\'")
	p = strings.ReplaceAll(p, "[", "\\[")
	p = strings.ReplaceAll(p, "]", "\\]")
	return p
}

// MediaInfo holds metadata about a media file.
type MediaInfo struct {
	Duration float64      `json:"duration"`
	Format   string       `json:"format"`
	Size     int64        `json:"size"`
	BitRate  string       `json:"bitRate"`
	Streams  []StreamInfo `json:"streams"`
}

// StreamInfo holds metadata about a single stream within a media file.
type StreamInfo struct {
	Index      int    `json:"index"`
	Type       string `json:"type"`
	Codec      string `json:"codec"`
	BitRate    string `json:"bitRate"`
	SampleRate string `json:"sampleRate,omitempty"`
	Channels   int    `json:"channels,omitempty"`
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
}

// ffprobeOutput represents the JSON output of ffprobe.
type ffprobeOutput struct {
	Format  ffprobeFormat   `json:"format"`
	Streams []ffprobeStream `json:"streams"`
}

type ffprobeFormat struct {
	FormatName     string `json:"format_name"`
	FormatLongName string `json:"format_long_name"`
	Duration       string `json:"duration"`
	Size           string `json:"size"`
	BitRate        string `json:"bit_rate"`
}

type ffprobeStream struct {
	Index         int    `json:"index"`
	CodecType     string `json:"codec_type"`
	CodecName     string `json:"codec_name"`
	BitRate       string `json:"bit_rate"`
	SampleRate    string `json:"sample_rate"`
	Channels      int    `json:"channels"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	CodecLongName string `json:"codec_long_name"`
}

// findFFmpeg resolves the ffmpeg binary path dynamically across Windows, Linux, and Docker.
func findFFmpeg() string {
	if envPath := os.Getenv("FFMPEG_PATH"); envPath != "" {
		return envPath
	}
	if path, err := exec.LookPath("ffmpeg"); err == nil {
		return path
	}
	return "ffmpeg"
}

// findFFprobe resolves the ffprobe binary path dynamically across Windows, Linux, and Docker.
func findFFprobe() string {
	if envPath := os.Getenv("FFPROBE_PATH"); envPath != "" {
		return envPath
	}
	if path, err := exec.LookPath("ffprobe"); err == nil {
		return path
	}
	return "ffprobe"
}

// ProbeFile runs ffprobe on the given file and returns parsed MediaInfo.
func ProbeFile(filePath string) (*MediaInfo, error) {
	filePath = cleanPath(filePath)
	ffprobe := findFFprobe()
	cmd := exec.Command(ffprobe,
		"-v", "error",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath,
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errStr := strings.TrimSpace(stderr.String())
		if errStr == "" {
			errStr = err.Error()
		}
		return nil, fmt.Errorf("ffprobe failed: %s", errStr)
	}

	var output ffprobeOutput
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %v", err)
	}

	info := &MediaInfo{
		Format:  output.Format.FormatName,
		BitRate: output.Format.BitRate,
	}

	if d, err := strconv.ParseFloat(output.Format.Duration, 64); err == nil {
		info.Duration = d
	}
	if s, err := strconv.ParseInt(output.Format.Size, 10, 64); err == nil {
		info.Size = s
	}

	for _, s := range output.Streams {
		stream := StreamInfo{
			Index:   s.Index,
			Codec:   s.CodecName,
			BitRate: s.BitRate,
		}
		switch s.CodecType {
		case "video":
			stream.Type = "video"
			stream.Width = s.Width
			stream.Height = s.Height
		case "audio":
			stream.Type = "audio"
			stream.SampleRate = s.SampleRate
			stream.Channels = s.Channels
		case "subtitle":
			stream.Type = "subtitle"
		default:
			stream.Type = s.CodecType
		}
		info.Streams = append(info.Streams, stream)
	}

	return info, nil
}

// timePattern matches FFmpeg progress lines like "time=HH:MM:SS.ms" or "time=00:00:01".
var timePattern = regexp.MustCompile(`time=(-?\d+:\d+:\d+(?:\.\d+)?)`)

// FFmpegLogHandler receives parsed progress (0-100) and raw log lines from FFmpeg's stderr in real-time.
type FFmpegLogHandler func(progress float64, logLine string)

// RunFFmpegContext executes FFmpeg with context cancellation, real-time stderr logging, and progress tracking.
func RunFFmpegContext(ctx context.Context, args []string, totalDuration float64, handler FFmpegLogHandler) error {
	ffmpeg := findFFmpeg()
	cmd := exec.CommandContext(ctx, ffmpeg, args...)

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %v", err)
	}

	var stderrLines []string
	scanner := bufio.NewScanner(stderrPipe)
	scanner.Split(scanFFmpegLines)

	currentProgress := 0.0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		stderrLines = append(stderrLines, line)

		if totalDuration > 0 {
			matches := timePattern.FindStringSubmatch(line)
			if len(matches) > 1 {
				currentTime := parseTimeToSeconds(matches[1])
				currentProgress = (currentTime / totalDuration) * 100
				if currentProgress < 0 {
					currentProgress = 0
				}
				if currentProgress > 100 {
					currentProgress = 100
				}
			}
		}

		if handler != nil {
			handler(currentProgress, line)
		}
	}

	if err := cmd.Wait(); err != nil {
		if ctx.Err() == context.Canceled {
			return fmt.Errorf("process was cancelled by client")
		}
		errMsg := strings.Join(lastN(stderrLines, 10), "\n")
		return fmt.Errorf("ffmpeg error: %v\n%s", err, errMsg)
	}

	return nil
}

// RunFFmpeg executes ffmpeg with the given args and a progress callback (backward compatible).
func RunFFmpeg(args []string, totalDuration float64, onProgress func(float64)) error {
	return RunFFmpegContext(context.Background(), args, totalDuration, func(p float64, _ string) {
		if onProgress != nil {
			onProgress(p)
		}
	})
}

// RunFFmpegPipe streams stdin directly to FFmpeg pipe:0 and stdout to an io.Writer pipe:1.
// This allows streaming uploads directly through FFmpeg with zero disk and bounded RAM.
func RunFFmpegPipe(ctx context.Context, args []string, stdin io.Reader, stdout io.Writer, handler FFmpegLogHandler) error {
	ffmpeg := findFFmpeg()
	cmd := exec.CommandContext(ctx, ffmpeg, args...)

	if stdin != nil {
		cmd.Stdin = stdin
	}
	if stdout != nil {
		cmd.Stdout = stdout
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg pipe: %v", err)
	}

	var stderrLines []string
	scanner := bufio.NewScanner(stderrPipe)
	scanner.Split(scanFFmpegLines)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		stderrLines = append(stderrLines, line)
		if handler != nil {
			handler(0, line)
		}
	}

	if err := cmd.Wait(); err != nil {
		if ctx.Err() == context.Canceled {
			return fmt.Errorf("pipe process was cancelled")
		}
		errMsg := strings.Join(lastN(stderrLines, 10), "\n")
		return fmt.Errorf("ffmpeg pipe error: %v\n%s", err, errMsg)
	}

	return nil
}

// parseTimeToSeconds converts "HH:MM:SS.ms", "MM:SS.ms", or "SS.ms" to total seconds.
func parseTimeToSeconds(timeStr string) float64 {
	timeStr = strings.TrimSpace(timeStr)
	parts := strings.Split(timeStr, ":")
	if len(parts) == 3 {
		h, _ := strconv.ParseFloat(parts[0], 64)
		m, _ := strconv.ParseFloat(parts[1], 64)
		s, _ := strconv.ParseFloat(parts[2], 64)
		return h*3600 + m*60 + s
	} else if len(parts) == 2 {
		m, _ := strconv.ParseFloat(parts[0], 64)
		s, _ := strconv.ParseFloat(parts[1], 64)
		return m*60 + s
	} else if len(parts) == 1 {
		s, _ := strconv.ParseFloat(parts[0], 64)
		return s
	}
	return 0
}

// scanFFmpegLines splits on \r or \n.
func scanFFmpegLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	for i := 0; i < len(data); i++ {
		if data[i] == '\n' || data[i] == '\r' {
			return i + 1, data[:i], nil
		}
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

// lastN returns the last n elements of a string slice.
func lastN(s []string, n int) []string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}
