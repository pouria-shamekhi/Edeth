package main

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
)

//go:embed frontend
var frontendFS embed.FS

func main() {
	port := 8420
	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil && p > 0 {
			port = p
		}
	}

	srv := NewServer(frontendFS, port)

	url := fmt.Sprintf("http://localhost:%d", port)
	fmt.Printf("Edeth is running at %s (PID: %d)\n", url, os.Getpid())

	// Auto-open browser in desktop mode (skip in headless/Docker environments)
	if os.Getenv("DOCKER") == "" && os.Getenv("HEADLESS") == "" {
		go func() {
			switch runtime.GOOS {
			case "windows":
				_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
			case "darwin":
				_ = exec.Command("open", url).Start()
			case "linux":
				_ = exec.Command("xdg-open", url).Start()
			}
		}()
	}

	if err := srv.Start(); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
