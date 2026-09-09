# How it works

## Overview

1. You choose a file (argument or picker).
2. `envia` validates it's a file (not a directory) and gets size.
3. Opens an HTTP server only on `127.0.0.1` with a port chosen by the system.
4. Calls `bore local PORT --to bore.pub` (default, 2-4s) or `cloudflared tunnel --url http://127.0.0.1:PORT` if `--provider cloudflare`.
5. Reads the process output and captures the URL (`bore.pub:PORT` or `https://*.trycloudflare.com`).
6. For `bore`, returns immediately (no health); for `cloudflare`, validates with a fast `GET /` (up to 3 retries, 2.5s each). If `bore` fails 3×, tries `serveo` (`ssh -R 80:localhost:PORT serveo.net` → `https://*.serveo.net`) and then `localhost.run` (`https://*.lhr.life`).
7. Shows the link with `bubbles/spinner` and QR code (`skip2/go-qrcode`, <5ms) while creating.
8. Serves the file with streaming and waits until `Ctrl+C`.

While the spinner shows “Creating link…”, DNS retries happen underneath without polluting the terminal.

## Under the hood

**HTTP (`net/http` pure, no frameworks):**
- `127.0.0.1:0` → random port, never `0.0.0.0` in 0.2.3.
- Endpoints: `GET /` (HTML), `GET /preview` (inline), `GET /download` (attachment), `GET /preview.css`, `GET /app.js`, `GET /qr` (QR PNG).
- `http.ServeContent` with `os.Open` → doesn't load 20GB into RAM, supports `Range: bytes=...` (`206 Partial Content`) for seek and resume.
- Headers: `X-Content-Type-Options: nosniff`, `Referrer-Policy: no-referrer`, `Accept-Ranges: bytes`.
- Only serves the chosen file — no `../` navigation.

**Page (`embed.FS`):**
- `cmd/envia/web/*` (`index.html`, `preview.css`, `app.js`) embedded via `//go:embed web/*`.
- `html/template` with escaping for spaces/accents/unicode.
- Preview: `image` → `<img>`, `video` → `<video controls>`, `audio` → `<audio>`, `pdf` → `<iframe>`, generic fallback. Pure CSS with `prefers-color-scheme`, 100% offline (no CDN).

**Tunnel (`os/exec`):**
- `exec.LookPath` checks `bore`/`cloudflared`/`ssh` in `PATH` (`bore` is default, bundle in `third-party/bore/` or next to `envia`).
- `exec.CommandContext` with `signal.NotifyContext` to kill children on `Ctrl+C`.
- Regex `bore\.pub:(\d+)`, `https://[a-zA-Z0-9-]+\.serveo\.net`, `https://[a-zA-Z0-9-]+\.lhr\.life` and `https://[a-zA-Z0-9-]+\.trycloudflare\.com` with ANSI strip.
- `bore`: `12s` for URL, no health (instant, 3 retries with `100ms` gap). `cloudflare`: `15s` for URL + health `2.5s` per try, 3 tries.

**UI (`charmbracelet/bubbletea` + `bubbles`):**
- Model with states `picking → loading → ready`. `list` for files, `spinner.Dot` for loading.
- QR cached on `tunnelReadyMsg` (generated once, not every frame).
- Fallback to plain mode (`app.Run`) when no TTY (pipes/CI).

**QR code:**
- `internal/qr/qr.go` with `skip2/go-qrcode` (Low recovery, 256x256 PNG for web, ASCII with half-blocks `▀▄█` for terminal, ~13 lines vs 30 before).

## Diagram

```
Your PC (envia)
  └─ file on disk
  └─ http://127.0.0.1:PORT (ServeContent)
  └─ bore local PORT --to bore.pub  ──┐
                                      │ tunnel
                                      ▼
                                   bore.pub:PORT
                                      │
                                      ▼
                                   Browser / phone (QR)
                                    → GET /preview (Range)
                                    → GET /download (attachment)
                                    → GET /qr (PNG)
```
