package main

import (
	"os"
	"path/filepath"
	"testing"
)

func resolveSamplePath(name string) string {
	samplePath := filepath.Join("samples", name)
	if _, err := os.Stat(samplePath); err == nil {
		return samplePath
	}
	return name
}

func TestWatermarkSRM(t *testing.T) {
	inFile := resolveSamplePath(`music_watermarked.mp3`)
	wmFile := resolveSamplePath(`wateraudio.srm`)
	outFile := `music_watermarked_test_out.mp3`

	if _, err := os.Stat(inFile); os.IsNotExist(err) {
		t.Skip("test file not found")
	}

	_ = os.Remove(outFile)

	s := &Server{tasks: make(map[string]*Task)}
	task := s.createTask()

	go func() {
		for u := range task.updates {
			t.Logf("[Progress %.1f%%] %s (%s)", u.Progress, u.Message, u.Status)
		}
	}()

	err := RemoveWatermark(inFile, wmFile, outFile, task)
	if err != nil {
		t.Fatalf("RemoveWatermark returned error: %v", err)
	}

	stat, err := os.Stat(outFile)
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}
	if stat.Size() < 100000 {
		t.Fatalf("Output file too small: %d bytes", stat.Size())
	}

	t.Logf("Success! Output file size: %d bytes", stat.Size())
	_ = os.Remove(outFile)
}

func TestWatermarkMP3(t *testing.T) {
	inFile := resolveSamplePath(`music_watermarked.mp3`)
	wmFile := resolveSamplePath(`AudioJungle watermark sound.mp3`)
	outFile := `music_watermarked_test_mp3_out.mp3`

	if _, err := os.Stat(inFile); os.IsNotExist(err) {
		t.Skip("test file not found")
	}

	_ = os.Remove(outFile)

	s := &Server{tasks: make(map[string]*Task)}
	task := s.createTask()

	go func() {
		for u := range task.updates {
			t.Logf("[Progress %.1f%%] %s (%s)", u.Progress, u.Message, u.Status)
		}
	}()

	err := RemoveWatermark(inFile, wmFile, outFile, task)
	if err != nil {
		t.Fatalf("RemoveWatermark returned error: %v", err)
	}

	stat, err := os.Stat(outFile)
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}
	if stat.Size() < 100000 {
		t.Fatalf("Output file too small: %d bytes", stat.Size())
	}

	t.Logf("Success! Output file size: %d bytes", stat.Size())
	_ = os.Remove(outFile)
}
