package httpserver

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Server serves a single file.
type Server struct {
	filePath    string
	fileName    string
	fileSize    int64
	mimeType    string
	previewType string
	modTime     int64 // unix not used, but for ServeContent modtime
	listener    net.Listener
	httpServer  *http.Server
	webFS       fs.FS
	tmpl        *template.Template
}

// New creates a new Server for the given file. webFS should be the embedded web files (or nil for tests).
func New(filePath string, webFS fs.FS) (*Server, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("arquivo não encontrado: %s: %w", filePath, err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("o caminho informado é um diretório. Escolha um arquivo")
	}

	abs, err := filepath.Abs(filePath)
	if err != nil {
		abs = filePath
	}

	ext := strings.ToLower(filepath.Ext(abs))
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		// try to detect some common types not in mime db
		switch ext {
		case ".md":
			mimeType = "text/markdown"
		case ".js":
			mimeType = "application/javascript"
		default:
			mimeType = "application/octet-stream"
		}
	}
	// Ensure mime without charset for binary?
	// Keep as is; ServeContent will use it.

	previewType := detectPreviewType(ext)

	// Load template if webFS available
	var tmpl *template.Template
	if webFS != nil {
		data, err := fs.ReadFile(webFS, "index.html")
		if err != nil {
			// try web/index.html
			data, err = fs.ReadFile(webFS, "web/index.html")
			if err != nil {
				return nil, fmt.Errorf("falha ao ler template: %w", err)
			}
		}
		tmpl, err = template.New("index").Parse(string(data))
		if err != nil {
			return nil, fmt.Errorf("falha ao parsear template: %w", err)
		}
	}

	s := &Server{
		filePath:    abs,
		fileName:    filepath.Base(abs),
		fileSize:    info.Size(),
		mimeType:    mimeType,
		previewType: previewType,
		webFS:       webFS,
		tmpl:        tmpl,
	}
	return s, nil
}

func detectPreviewType(ext string) string {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg":
		return "image"
	case ".mp4", ".webm", ".mov", ".m4v", ".ogv":
		return "video"
	case ".mp3", ".wav", ".flac", ".aac", ".ogg", ".opus", ".oga", ".m4a":
		return "audio"
	case ".pdf":
		return "pdf"
	default:
		return "generic"
	}
}

// Addr returns the listening address (e.g., 127.0.0.1:43821) after Start.
func (s *Server) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return ""
}

// URL returns http:// + Addr
func (s *Server) URL() string {
	return "http://" + s.Addr()
}

// Listen starts listening on 127.0.0.1:0 and starts serving in background.
// It returns the address. Caller should ensure Close is called.
func (s *Server) Listen() (string, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("falha ao iniciar listener: %w", err)
	}
	s.listener = ln

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/preview", s.handlePreview)
	mux.HandleFunc("/download", s.handleDownload)
	mux.HandleFunc("/preview.css", s.handleCSS)
	mux.HandleFunc("/app.js", s.handleJS)

	// Security headers middleware
	handler := securityHeaders(mux)

	s.httpServer = &http.Server{
		Handler: handler,
	}

	go func() {
		// Serve will return ErrServerClosed on Close
		_ = s.httpServer.Serve(ln)
	}()

	return ln.Addr().String(), nil
}

// Close shuts down the server.
func (s *Server) Close() error {
	if s.httpServer != nil {
		_ = s.httpServer.Close()
	}
	if s.listener != nil {
		_ = s.listener.Close()
	}
	return nil
}

// Shutdown gracefully.
func (s *Server) Shutdown() error {
	if s.httpServer != nil {
		return s.httpServer.Close()
	}
	return nil
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		// Only allow our endpoints; return 404 for unknown paths except /
		next.ServeHTTP(w, r)
	})
}

type pageData struct {
	FileName    string
	FileSize    string
	SizeBytes   int64
	Ext         string
	MimeType    string
	PreviewType string
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if s.tmpl != nil {
		data := pageData{
			FileName:    s.fileName,
			FileSize:    humanSize(s.fileSize),
			SizeBytes:   s.fileSize,
			Ext:         filepath.Ext(s.fileName),
			MimeType:    s.mimeType,
			PreviewType: s.previewType,
		}
		if err := s.tmpl.Execute(w, data); err != nil {
			http.Error(w, "erro ao renderizar", http.StatusInternalServerError)
		}
		return
	}
	// Fallback minimal HTML if no template (for tests without webFS)
	_, _ = io.WriteString(w, fallbackHTML(s.fileName, humanSize(s.fileSize)))
}

func (s *Server) handlePreview(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/preview" {
		http.NotFound(w, r)
		return
	}
	s.serveFile(w, r, false)
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/download" {
		http.NotFound(w, r)
		return
	}
	s.serveFile(w, r, true)
}

func (s *Server) serveFile(w http.ResponseWriter, r *http.Request, asAttachment bool) {
	f, err := os.Open(s.filePath)
	if err != nil {
		http.Error(w, "arquivo não encontrado", http.StatusNotFound)
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		http.Error(w, "erro ao ler arquivo", http.StatusInternalServerError)
		return
	}

	// Security: only serve the selected file, with correct mime
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Accept-Ranges", "bytes")

	if asAttachment {
		// Content-Disposition with filename* for unicode
		escaped := template.JSEscapeString(s.fileName)
		_ = escaped // keep for later
		// Use RFC 5987 for utf-8
		// Use both filename and filename*
		// Need to escape quotes
		safeName := strings.ReplaceAll(s.fileName, `"`, `_`)
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, safeName, urlPathEscape(s.fileName)))
		// For generic download, force octet-stream? But spec says preserve mime but disposition attachment will trigger download.
		// We keep original mime for consistency, but browsers will download due to disposition.
	} else {
		// For preview, set Content-Disposition inline (optional) and correct mime
		w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, strings.ReplaceAll(s.fileName, `"`, `_`)))
	}

	// ServeContent handles Range, If-Modified-Since, Content-Type, etc.
	// It requires ReadSeeker.
	http.ServeContent(w, r, s.fileName, info.ModTime(), f)
}

func urlPathEscape(s string) string {
	return url.PathEscape(s)
}

func humanSize(b int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case b >= GB:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(GB))
	case b >= MB:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(MB))
	case b >= KB:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(KB))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func fallbackHTML(name, size string) string {
	return fmt.Sprintf(`<!doctype html><html><head><meta charset="utf-8"><title>envia</title></head><body><h1>%s</h1><p>%s</p><a href="/download">Baixar</a><div><a href="/preview">Preview</a></div></body></html>`, template.HTMLEscapeString(name), template.HTMLEscapeString(size))
}

func (s *Server) handleCSS(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/preview.css" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	if s.webFS != nil {
		for _, p := range []string{"preview.css", "web/preview.css"} {
			if data, err := fs.ReadFile(s.webFS, p); err == nil {
				_, _ = w.Write(data)
				return
			}
		}
	}
	http.NotFound(w, r)
}

func (s *Server) handleJS(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/app.js" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	if s.webFS != nil {
		for _, p := range []string{"app.js", "web/app.js"} {
			if data, err := fs.ReadFile(s.webFS, p); err == nil {
				_, _ = w.Write(data)
				return
			}
		}
	}
	http.NotFound(w, r)
}

// For testing: create handler without listening
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/preview", s.handlePreview)
	mux.HandleFunc("/download", s.handleDownload)
	mux.HandleFunc("/preview.css", s.handleCSS)
	mux.HandleFunc("/app.js", s.handleJS)
	return securityHeaders(mux)
}
