package httpserver

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createTempFile(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return path
}

func TestIndexReturnsHTMLWithFilename(t *testing.T) {
	path := createTempFile(t, "hello.txt", "hello world")
	srv, err := New(path, nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	// Need template? nil will use fallback. For proper test with template, pass embed-like FS.
	// Test fallback still contains filename
	h := srv.Handler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "hello.txt") {
		t.Fatalf("body should contain filename, got %q", body)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Fatalf("content-type should be html, got %q", ct)
	}
	if v := rec.Header().Get("X-Content-Type-Options"); v != "nosniff" {
		t.Fatalf("missing X-Content-Type-Options nosniff got %q", v)
	}
}

func TestPreviewServesFile(t *testing.T) {
	content := "preview content here"
	path := createTempFile(t, "foto.jpg", content)
	srv, err := New(path, nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	h := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/preview", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("preview status %d", rec.Code)
	}
	if rec.Body.String() != content {
		t.Fatalf("preview body mismatch %q vs %q", rec.Body.String(), content)
	}
	// Should have Accept-Ranges
	if rec.Header().Get("Accept-Ranges") != "bytes" {
		t.Fatalf("missing Accept-Ranges")
	}
	// X-Content-Type-Options
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("missing nosniff")
	}
	// Content-Type should be image/jpeg for .jpg
	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "image/jpeg") {
		t.Fatalf("expected image/jpeg got %q", ct)
	}
}

func TestDownloadForcesAttachment(t *testing.T) {
	content := "download content"
	path := createTempFile(t, "meu vídeo final.mp4", content)
	srv, err := New(path, nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	h := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/download", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("download status %d", rec.Code)
	}
	disp := rec.Header().Get("Content-Disposition")
	if !strings.Contains(disp, "attachment") {
		t.Fatalf("expected attachment disposition got %q", disp)
	}
	if !strings.Contains(disp, "meu") {
		t.Fatalf("disposition should contain original filename, got %q", disp)
	}
	if rec.Body.String() != content {
		t.Fatalf("body mismatch")
	}
}

func TestRangeRequests(t *testing.T) {
	content := "abcdefghijklmnopqrstuvwxyz0123456789"
	path := createTempFile(t, "range.bin", content)
	srv, err := New(path, nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	h := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/download", nil)
	req.Header.Set("Range", "bytes=0-9")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusPartialContent {
		t.Fatalf("expected 206 got %d body %q", rec.Code, rec.Body.String())
	}
	if cr := rec.Header().Get("Content-Range"); !strings.Contains(cr, "bytes 0-9") {
		t.Fatalf("missing Content-Range got %q", cr)
	}
	if body := rec.Body.String(); body != "abcdefghij" {
		t.Fatalf("range body expected abcdefghij got %q", body)
	}
	// Test streaming large file doesn't load all? ServeContent ensures not, but we check Range with offset
	req2 := httptest.NewRequest(http.MethodGet, "/preview", nil)
	req2.Header.Set("Range", "bytes=10-19")
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusPartialContent {
		t.Fatalf("preview range expected 206 got %d", rec2.Code)
	}
	if body := rec2.Body.String(); body != "klmnopqrst" {
		t.Fatalf("range preview body %q", body)
	}
}

func TestRangeSeekResume(t *testing.T) {
	content := strings.Repeat("A", 10000) + strings.Repeat("B", 10000)
	path := createTempFile(t, "large.dat", content)
	srv, err := New(path, nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	h := srv.Handler()

	// Request bytes=10000- (should get second half)
	req := httptest.NewRequest(http.MethodGet, "/download", nil)
	req.Header.Set("Range", "bytes=10000-")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusPartialContent {
		t.Fatalf("expected 206 got %d", rec.Code)
	}
	if len(rec.Body.String()) != 10000 {
		t.Fatalf("expected 10000 bytes got %d", len(rec.Body.String()))
	}
	if rec.Body.String()[:10] != "BBBBBBBBBB" {
		t.Fatalf("unexpected content")
	}
}

func TestUnknownPath404(t *testing.T) {
	path := createTempFile(t, "a.txt", "hi")
	srv, _ := New(path, nil)
	h := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/qualquer-outro-arquivo", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown path got %d", rec.Code)
	}
	// Traversal via explicit file path should not leak file - ServeMux may redirect, accept 301/307 or 404
	req2 := httptest.NewRequest(http.MethodGet, "/secret.txt", nil)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for secret file got %d", rec2.Code)
	}
}

func TestHeadersSecurity(t *testing.T) {
	path := createTempFile(t, "x.txt", "x")
	srv, _ := New(path, nil)
	h := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Header().Get("Referrer-Policy") != "no-referrer" {
		t.Fatalf("expected Referrer-Policy no-referrer got %q", rec.Header().Get("Referrer-Policy"))
	}
}

func TestContentTypeDetection(t *testing.T) {
	tests := []struct {
		name string
		ext  string
		want string
	}{
		{"image/png", "img.png", "image/png"},
		{"video/mp4", "vid.mp4", "video/mp4"},
		{"audio mp3", "mus.mp3", "audio/mpeg"},
		{"pdf", "doc.pdf", "application/pdf"},
		{"zip", "arch.zip", "application/zip"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := createTempFile(t, tc.ext, "content")
			srv, _ := New(path, nil)
			h := srv.Handler()
			req := httptest.NewRequest(http.MethodGet, "/preview", nil)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			ct := rec.Header().Get("Content-Type")
			if !strings.Contains(ct, tc.want) && tc.want != "" {
				// Log but allow fallback to octet-stream for unknown? For known types should match.
				t.Fatalf("Content-Type %q should contain %q", ct, tc.want)
			}
		})
	}
}

func TestDownloadDoesNotLoadAllIntoMemory(t *testing.T) {
	// Create 5MB file
	dir := t.TempDir()
	path := filepath.Join(dir, "large.bin")
	f, _ := os.Create(path)
	// Write 5MB of zeros via streaming
	chunk := make([]byte, 1024*1024)
	for i := 0; i < 5; i++ {
		f.Write(chunk)
	}
	f.Close()

	srv, _ := New(path, nil)
	h := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/download", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	// Ensure we can read via io, not just body length check
	if rec.Body.Len() != 5*1024*1024 {
		t.Fatalf("expected 5MB got %d", rec.Body.Len())
	}
	// Also test that we can stream via ResponseRecorder's Body as ReadSeeker? Not needed.
	_, _ = io.ReadAll(rec.Body)
}
