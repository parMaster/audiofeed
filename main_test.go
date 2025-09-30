package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFromMediaFolder(t *testing.T) {
	// Create a temporary directory with test files
	tmpDir := t.TempDir()

	// Create some test subdirectories
	os.MkdirAll(filepath.Join(tmpDir, "title1"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "title2"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, ".hidden"), 0755) // Should be skipped

	server := &feedServer{Options: Options{MediaFolder: tmpDir}}
	titles, err := server.fromMediaFolder(tmpDir)

	if err != nil {
		t.Fatalf("fromMediaFolder returned error: %v", err)
	}

	if len(titles) != 2 {
		t.Errorf("Expected 2 titles, got %d", len(titles))
	}

	// Check that .hidden is not included
	for _, title := range titles {
		if strings.HasPrefix(title, ".") {
			t.Errorf("Hidden directory should not be included: %s", title)
		}
	}
}

func TestReadTitle(t *testing.T) {
	// Create a temporary directory with test files
	tmpDir := t.TempDir()
	titleDir := filepath.Join(tmpDir, "test-title")
	os.MkdirAll(titleDir, 0755)

	// Create test files
	os.WriteFile(filepath.Join(titleDir, "chapter1.mp3"), []byte("fake audio"), 0644)
	os.WriteFile(filepath.Join(titleDir, "chapter2.m4a"), []byte("fake audio"), 0644)
	os.WriteFile(filepath.Join(titleDir, "chapter3.m4b"), []byte("fake audio"), 0644)
	os.WriteFile(filepath.Join(titleDir, "cover.jpg"), []byte("fake image"), 0644)
	os.WriteFile(filepath.Join(titleDir, "readme.txt"), []byte("ignore this"), 0644)

	server := &feedServer{}
	chapters, coverPath, err := server.readTitle(titleDir)

	if err != nil {
		t.Fatalf("readTitle returned error: %v", err)
	}

	if len(chapters) != 3 {
		t.Errorf("Expected 3 chapters, got %d", len(chapters))
	}

	if coverPath == "" {
		t.Error("Expected cover path to be set")
	}

	if !strings.Contains(coverPath, "cover.jpg") {
		t.Errorf("Expected cover path to contain 'cover.jpg', got %s", coverPath)
	}
}

func TestReadTitleNonExistent(t *testing.T) {
	server := &feedServer{}
	_, _, err := server.readTitle("/nonexistent/path")

	if err == nil {
		t.Error("Expected error for non-existent path")
	}
}

func TestFilesOnlyMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	middleware := filesOnly(handler)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{"File path", "/audio/test.mp3", http.StatusOK},
		{"Directory path", "/audio/", http.StatusNotFound},
		{"Root path", "/", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			middleware.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestInfoHandler(t *testing.T) {
	server := &feedServer{}
	req := httptest.NewRequest(http.MethodGet, "/info", nil)
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()

	server.info(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if result["UserAgent"] != "test-agent" {
		t.Errorf("Expected UserAgent 'test-agent', got %s", result["UserAgent"])
	}
}

func TestStylesheetHandler(t *testing.T) {
	server := &feedServer{}
	req := httptest.NewRequest(http.MethodGet, "/feed.xsl", nil)
	w := httptest.NewRecorder()

	server.stylesheet(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/xsl") {
		t.Errorf("Expected Content-Type to contain text/xsl, got %s", contentType)
	}
}

func TestIndexHandler(t *testing.T) {
	// Create a temporary directory with test titles
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "title1"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "title2"), 0755)

	tests := []struct {
		name           string
		accessCode     string
		requestPath    string
		expectedStatus int
	}{
		{"No access code required", "", "/index", http.StatusOK},
		{"Access code required and provided", "secret", "/index/secret", http.StatusOK},
		{"Access code required but wrong", "secret", "/index/wrong", http.StatusForbidden},
		{"Access code required but not provided", "secret", "/index", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &feedServer{Options: Options{
				MediaFolder: tmpDir,
				AccessCode:  tt.accessCode,
			}}

			req := httptest.NewRequest(http.MethodGet, tt.requestPath, nil)

			// Set path value if code is in path
			if strings.Contains(tt.requestPath, "/index/") {
				code := strings.TrimPrefix(tt.requestPath, "/index/")
				req.SetPathValue("code", code)
			}

			w := httptest.NewRecorder()
			server.index(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestDisplayTitleHandler(t *testing.T) {
	// Create a temporary directory with test title
	tmpDir := t.TempDir()
	titleDir := filepath.Join(tmpDir, "test-title")
	os.MkdirAll(titleDir, 0755)
	os.WriteFile(filepath.Join(titleDir, "chapter1.mp3"), []byte("fake audio"), 0644)

	server := &feedServer{Options: Options{MediaFolder: tmpDir}}
	req := httptest.NewRequest(http.MethodGet, "/title/test-title", nil)
	req.SetPathValue("title", "test-title")
	w := httptest.NewRecorder()

	server.displayTitle(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/xml") {
		t.Errorf("Expected Content-Type to contain text/xml, got %s", contentType)
	}

	body, _ := io.ReadAll(w.Body)
	if !strings.Contains(string(body), "chapter1.mp3") {
		t.Error("Expected response to contain chapter1.mp3")
	}
}

func TestDisplayTitleHandlerNonExistent(t *testing.T) {
	tmpDir := t.TempDir()

	server := &feedServer{Options: Options{MediaFolder: tmpDir}}
	req := httptest.NewRequest(http.MethodGet, "/title/nonexistent", nil)
	req.SetPathValue("title", "nonexistent")
	w := httptest.NewRecorder()

	server.displayTitle(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestNewFeedServer(t *testing.T) {
	opts := Options{
		Port:        8080,
		MediaFolder: "/test/path",
		AccessCode:  "secret",
	}

	server := NewFeedServer(opts)

	if server == nil {
		t.Fatal("NewFeedServer returned nil")
	}

	if server.Port != opts.Port {
		t.Errorf("Expected Port %d, got %d", opts.Port, server.Port)
	}

	if server.MediaFolder != opts.MediaFolder {
		t.Errorf("Expected MediaFolder %s, got %s", opts.MediaFolder, server.MediaFolder)
	}

	if server.AccessCode != opts.AccessCode {
		t.Errorf("Expected AccessCode %s, got %s", opts.AccessCode, server.AccessCode)
	}
}
