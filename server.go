package main

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"sync"
)

// TaskUpdate represents a progress update and real-time log for a task.
type TaskUpdate struct {
	Progress   float64 `json:"progress"`
	Status     string  `json:"status"`
	Message    string  `json:"message"`
	Log        string  `json:"log,omitempty"`
	OutputFile string  `json:"outputFile,omitempty"`
}

// Task represents an async processing task with context cancellation.
type Task struct {
	ID         string  `json:"id"`
	Status     string  `json:"status"`
	Progress   float64 `json:"progress"`
	Message    string  `json:"message"`
	OutputFile string  `json:"outputFile,omitempty"`

	ctx     context.Context
	cancel  context.CancelFunc
	done    chan struct{}
	updates chan TaskUpdate
}

// Server is the main HTTP server for Edeth.
type Server struct {
	port       int
	mux        *http.ServeMux
	tasks      map[string]*Task
	tasksMu    sync.RWMutex
	frontendFS embed.FS
}

// NewServer creates a new Server instance.
func NewServer(frontendFS embed.FS, port int) *Server {
	s := &Server{
		port:       port,
		mux:        http.NewServeMux(),
		tasks:      make(map[string]*Task),
		frontendFS: frontendFS,
	}
	s.setupRoutes()
	return s
}

// setupRoutes registers all API routes and serves the embedded frontend.
func (s *Server) setupRoutes() {
	// API routes
	s.mux.HandleFunc("/api/upload", s.handleUpload)
	s.mux.HandleFunc("/api/stream", s.handleStream)
	s.mux.HandleFunc("/api/browse", s.handleBrowse)
	s.mux.HandleFunc("/api/drives", s.handleDrives)
	s.mux.HandleFunc("/api/media-info", s.handleMediaInfo)
	s.mux.HandleFunc("/api/watermark/remove", s.handleWatermarkRemove)
	s.mux.HandleFunc("/api/subtitle/add", s.handleSubtitleAdd)
	s.mux.HandleFunc("/api/trim", s.handleTrim)
	s.mux.HandleFunc("/api/audio/extract", s.handleAudioExtract)
	s.mux.HandleFunc("/api/audio/merge", s.handleAudioMerge)
	s.mux.HandleFunc("/api/video/compress", s.handleVideoCompress)
	s.mux.HandleFunc("/api/progress/", s.handleProgress)

	// Serve embedded frontend files at "/" with no-cache headers to prevent browser caching stale assets
	frontendSub, err := fs.Sub(s.frontendFS, "frontend")
	if err != nil {
		panic(fmt.Sprintf("failed to create sub filesystem: %v", err))
	}
	fileServer := http.FileServer(http.FS(frontendSub))
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		fileServer.ServeHTTP(w, r)
	})
}

// Start begins listening and serving HTTP requests.
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	return http.ListenAndServe(addr, s.mux)
}

// createTask generates a unique task ID, stores the task, and returns it.
func (s *Server) createTask() *Task {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	id := hex.EncodeToString(b)

	ctx, cancel := context.WithCancel(context.Background())

	task := &Task{
		ID:       id,
		Status:   "pending",
		Progress: 0,
		ctx:      ctx,
		cancel:   cancel,
		done:     make(chan struct{}),
		updates:  make(chan TaskUpdate, 128),
	}

	s.tasksMu.Lock()
	s.tasks[id] = task
	s.tasksMu.Unlock()

	return task
}

// getTask retrieves a task by ID.
func (s *Server) getTask(id string) *Task {
	s.tasksMu.RLock()
	defer s.tasksMu.RUnlock()
	return s.tasks[id]
}

// writeJSON writes a JSON response with the given status code and data.
func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// writeError writes a JSON error response.
func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	s.writeJSON(w, status, map[string]string{"error": message})
}

// sendTaskUpdate sends a progress update to the task's channel.
func sendTaskUpdate(task *Task, progress float64, status, message string) {
	sendTaskLog(task, progress, status, message, "")
}

// sendTaskLog sends a progress update along with real-time console/log text.
func sendTaskLog(task *Task, progress float64, status, message, logLine string) {
	if task == nil {
		return
	}
	task.Progress = progress
	task.Status = status
	if message != "" {
		task.Message = message
	}
	select {
	case task.updates <- TaskUpdate{
		Progress: progress,
		Status:   status,
		Message:  task.Message,
		Log:      logLine,
	}:
	default:
	}
}

// completeTask marks a task as complete and closes its channels.
func completeTask(task *Task, outputFile string) {
	if task == nil {
		return
	}
	task.Status = "complete"
	task.Progress = 100
	task.OutputFile = outputFile
	task.Message = "Complete"
	select {
	case task.updates <- TaskUpdate{
		Progress:   100,
		Status:     "complete",
		Message:    "Complete",
		OutputFile: outputFile,
	}:
	default:
	}
	if task.cancel != nil {
		task.cancel()
	}
	close(task.done)
}

// failTask marks a task as failed and closes its channels.
func failTask(task *Task, message string) {
	if task == nil {
		return
	}
	task.Status = "error"
	task.Message = message
	select {
	case task.updates <- TaskUpdate{
		Progress: task.Progress,
		Status:   "error",
		Message:  message,
		Log:      "Error: " + message,
	}:
	default:
	}
	if task.cancel != nil {
		task.cancel()
	}
	close(task.done)
}

// handleProgress is an SSE endpoint at /api/progress/{taskId}.
func (s *Server) handleProgress(w http.ResponseWriter, r *http.Request) {
	// Extract taskId from URL path: /api/progress/{taskId}
	taskID := r.URL.Path[len("/api/progress/"):]
	if taskID == "" {
		s.writeError(w, http.StatusBadRequest, "missing task ID")
		return
	}

	task := s.getTask(taskID)
	if task == nil {
		s.writeError(w, http.StatusNotFound, "task not found")
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if !ok {
		s.writeError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	flusher.Flush()

	ctx := r.Context()

	for {
		select {
		case <-ctx.Done():
			return
		case update, ok := <-task.updates:
			if !ok {
				return
			}
			data, _ := json.Marshal(update)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
			if update.Status == "complete" || update.Status == "error" {
				return
			}
		case <-task.done:
			// Drain remaining updates
			for {
				select {
				case update := <-task.updates:
					data, _ := json.Marshal(update)
					fmt.Fprintf(w, "data: %s\n\n", data)
					flusher.Flush()
				default:
					return
				}
			}
		}
	}
}
