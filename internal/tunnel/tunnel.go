package tunnel

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// Provider represents tunnel provider.
type Provider string

const (
	ProviderAuto       Provider = "auto"
	ProviderCloudflare Provider = "cloudflare"
	ProviderBore       Provider = "bore"
)

// Manager manages a tunnel process.
type Manager struct {
	cmd      *exec.Cmd
	provider Provider
	url      string
}

// Config for starting tunnel.
type Config struct {
	Addr     string   // e.g., 127.0.0.1:43821
	Provider Provider // auto, cloudflare, bore
}

// Regexes for URL discovery

var (
	cloudflareRegex = regexp.MustCompile(`https://[a-zA-Z0-9-]+\.trycloudflare\.com`)
	// bore output contains "bore.pub:12345" - may have ansi codes
	boreRegex = regexp.MustCompile(`bore\.pub:(\d+)`)
	ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)
)

// ParseCloudflareURL extracts trycloudflare URL from text.
func ParseCloudflareURL(text string) string {
	clean := ansiRegex.ReplaceAllString(text, "")
	m := cloudflareRegex.FindString(clean)
	return m
}

// ParseBoreURL extracts bore.pub URL from text and returns http:// URL.
func ParseBoreURL(text string) string {
	clean := ansiRegex.ReplaceAllString(text, "")
	m := boreRegex.FindString(clean)
	if m == "" {
		return ""
	}
	// m is bore.pub:PORT, return http://bore.pub:PORT
	return "http://" + m
}

func stripAnsi(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// Start starts the tunnel and returns public URL. It blocks until URL is found or timeout/context cancels.
// Timeout default 30s.
func (m *Manager) Start(ctx context.Context, cfg Config) (string, error) {
	return m.StartWithRetry(ctx, cfg)
}

// StartWithRetry starts tunnel with health validation and quick retries.
// Para auto: 3 tentativas cloudflare (health 2.2s cada) depois fallback bore.
// Para cloudflare explícito: 3 tentativas. Timeouts curtos pois retry subsequente já resolve DNS.
func (m *Manager) StartWithRetry(ctx context.Context, cfg Config) (string, error) {
	provider := cfg.Provider
	if provider == "" {
		provider = ProviderAuto
	}

	if provider == ProviderAuto {
		var lastErr error
		for i := 0; i < 3; i++ {
			if ctx.Err() != nil {
				return "", ctx.Err()
			}
			url, err := m.startCloudflare(ctx, cfg.Addr)
			if err != nil {
				lastErr = err
				if strings.Contains(err.Error(), "não encontrado") {
					break
				}
				select {
				case <-time.After(100 * time.Millisecond):
				case <-ctx.Done():
					return "", ctx.Err()
				}
				continue
			}
			m.provider = ProviderCloudflare
			m.url = url
			if err := waitHealthy(ctx, url); err == nil {
				return url, nil
			} else {
				lastErr = err
				_ = m.Close()
				select {
				case <-time.After(100 * time.Millisecond):
				case <-ctx.Done():
					return "", ctx.Err()
				}
				continue
			}
		}
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		// fallback bore (sem health, bore é rápido)
		url2, err2 := m.startBore(ctx, cfg.Addr)
		if err2 == nil {
			m.provider = ProviderBore
			m.url = url2
			return url2, nil
		}
		return "", fmt.Errorf("não foi possível criar um túnel público.\nVerifique se cloudflared ou bore está instalado e se sua conexão com a internet está funcionando: %v; bore: %v", lastErr, err2)
	}

	if provider == ProviderCloudflare {
		var lastErr error
		for i := 0; i < 3; i++ {
			if ctx.Err() != nil {
				return "", ctx.Err()
			}
			url, err := m.startCloudflare(ctx, cfg.Addr)
			if err != nil {
				lastErr = err
				select {
				case <-time.After(100 * time.Millisecond):
				case <-ctx.Done():
					return "", ctx.Err()
				}
				continue
			}
			m.provider = ProviderCloudflare
			m.url = url
			if err := waitHealthy(ctx, url); err == nil {
				return url, nil
			} else {
				lastErr = err
				_ = m.Close()
				select {
				case <-time.After(100 * time.Millisecond):
				case <-ctx.Done():
					return "", ctx.Err()
				}
			}
		}
		return "", fmt.Errorf("cloudflare indisponível após 3 tentativas: %v", lastErr)
	}

	if provider == ProviderBore {
		url, err := m.startBore(ctx, cfg.Addr)
		if err != nil {
			return "", err
		}
		m.provider = ProviderBore
		m.url = url
		return url, nil
	}

	return "", fmt.Errorf("provider desconhecido: %s", provider)
}

func waitHealthy(ctx context.Context, publicURL string) error {
	client := &http.Client{Timeout: 1500 * time.Millisecond}
	deadline := time.Now().Add(4000 * time.Millisecond)
	target := publicURL
	if !strings.HasSuffix(target, "/") {
		target += "/"
	}
	for time.Now().Before(deadline) {
			if ctx.Err() != nil {
			return ctx.Err()
		}
		req, err := http.NewRequestWithContext(context.Background(), "GET", target, nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			if resp.StatusCode < 500 {
					return nil
			}
		}
		select {
		case <-time.After(200 * time.Millisecond):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
return fmt.Errorf("health check falhou para %s", publicURL)
}

func (m *Manager) startCloudflare(ctx context.Context, addr string) (string, error) {
	if _, err := exec.LookPath("cloudflared"); err != nil {
		return "", fmt.Errorf("cloudflared não encontrado")
	}
	// Ensure addr has http:// prefix
	urlArg := addr
	if !strings.HasPrefix(urlArg, "http://") && !strings.HasPrefix(urlArg, "https://") {
		urlArg = "http://" + urlArg
	}
	// cloudflared tunnel --url http://127.0.0.1:PORT
	cmd := exec.CommandContext(ctx, "cloudflared", "tunnel", "--url", urlArg)
	// cloudflared logs to stderr and stdout, capture both
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("falha ao criar stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("falha ao criar stderr pipe: %w", err)
	}

	// Combine readers
	combined := io.MultiReader(stdout, stderr) // Note: MultiReader after pipes started? We need to start cmd first and read merged via both. But MultiReader will read stdout then stderr sequentially, not interleaved. Better to use StdoutPipe+StderrPipe concurrent reading via goroutine.
	// Alternative: set cmd.Stdout and Stderr to pipes and read concurrently.
	// Let's do proper: use io.Pipe for combined via goroutine.
	_ = combined // placeholder to avoid unused, we will implement properly below

	// Simpler: capture combined output via combined pipe using Stdout+Stderr to same pipe via cmd setup? But exec doesn't support merging directly, we need to read both pipes concurrently.
	// We'll start cmd, then read from both pipes concurrently via channels.
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("falha ao iniciar cloudflared: %w", err)
	}
	m.cmd = cmd

	// Channel for url
	urlCh := make(chan string, 1)
	errCh := make(chan error, 1)

	go func() {
		// Reader for stdout
		go func() {
			scanner := bufio.NewScanner(stdout)
			// Increase buffer for long lines
			buf := make([]byte, 0, 64*1024)
			scanner.Buffer(buf, 1024*1024)
			for scanner.Scan() {
				line := scanner.Text()
				clean := stripAnsi(line)
				if u := cloudflareRegex.FindString(clean); u != "" {
					select {
					case urlCh <- u:
					default:
					}
					return
				}
			}
		}()
		// Reader for stderr
		scanner := bufio.NewScanner(stderr)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			clean := stripAnsi(line)
			if u := cloudflareRegex.FindString(clean); u != "" {
				select {
				case urlCh <- u:
				default:
				}
				return
			}
		}
		// If we exit without finding url and process exited, send error
		// Wait a bit for process to exit
	}()

	// Also monitor process exit
	go func() {
		err := cmd.Wait()
		// If we haven't found URL and process exited, report
		select {
		case errCh <- fmt.Errorf("cloudflared encerrou sem gerar URL: %w", err):
		default:
		}
	}()

	timeout := 15 * time.Second
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case u := <-urlCh:
		return u, nil
	case err := <-errCh:
		_ = m.Close()
		return "", err
	case <-timer.C:
		_ = m.Close()
		return "", fmt.Errorf("timeout ao aguardar URL do cloudflared (15s)")
	case <-ctx.Done():
		_ = m.Close()
		return "", ctx.Err()
	}
}

func (m *Manager) startBore(ctx context.Context, addr string) (string, error) {
	if _, err := exec.LookPath("bore"); err != nil {
		return "", fmt.Errorf("bore não encontrado")
	}
	// addr is 127.0.0.1:PORT, need just PORT
	port := addr
	if strings.Contains(addr, ":") {
		parts := strings.Split(addr, ":")
		port = parts[len(parts)-1]
	}
	// bore local PORT --to bore.pub
	cmd := exec.CommandContext(ctx, "bore", "local", port, "--to", "bore.pub")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("falha ao criar stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("falha ao criar stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("falha ao iniciar bore: %w", err)
	}
	m.cmd = cmd

	urlCh := make(chan string, 1)
	errCh := make(chan error, 1)

	go func() {
		go func() {
			scanner := bufio.NewScanner(stdout)
			buf := make([]byte, 0, 64*1024)
			scanner.Buffer(buf, 1024*1024)
			for scanner.Scan() {
				line := scanner.Text()
				clean := stripAnsi(line)
				if m := boreRegex.FindString(clean); m != "" {
					select {
					case urlCh <- "http://" + m:
					default:
					}
					return
				}
			}
		}()
		scanner := bufio.NewScanner(stderr)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			clean := stripAnsi(line)
			if m := boreRegex.FindString(clean); m != "" {
				select {
				case urlCh <- "http://" + m:
				default:
				}
				return
			}
		}
	}()

	go func() {
		err := cmd.Wait()
		select {
		case errCh <- fmt.Errorf("bore encerrou sem gerar URL: %w", err):
		default:
		}
	}()

	timeout := 12 * time.Second
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case u := <-urlCh:
		return u, nil
	case err := <-errCh:
		_ = m.Close()
		return "", err
	case <-timer.C:
		_ = m.Close()
		return "", fmt.Errorf("timeout ao aguardar URL do bore (12s)")
	case <-ctx.Done():
		_ = m.Close()
		return "", ctx.Err()
	}
}

// Close terminates the tunnel process.
func (m *Manager) Close() error {
	if m.cmd != nil && m.cmd.Process != nil {
		_ = m.cmd.Process.Kill()
		// Do not Wait here if a goroutine is already waiting (avoid double Wait deadlock).
		// Give kernel a moment to reap, but don't block.
		// The Wait goroutine in startCloudflare/startBore will reap.
		m.cmd = nil
	}
	return nil
}

// Provider returns which provider is active.
func (m *Manager) Provider() Provider {
	return m.provider
}

// URL returns current public URL.
func (m *Manager) URL() string {
	return m.url
}
