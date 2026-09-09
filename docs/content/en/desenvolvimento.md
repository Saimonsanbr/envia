# Development

```bash
make run      # go run ./cmd/envia
make build    # builds bin/envia
make test     # go test ./...
make fmt      # gofmt -w .
make clean    # removes bin/
```

Tests cover:

- `internal/httpserver`: page, `Content-Type`, `Content-Disposition: attachment`, `Range 206`, headers, large file via streaming, `/qr` PNG
- `internal/tunnel`: URL parsing (`bore.pub`, `serveo.net`, `lhr.life`, `trycloudflare.com`) without network, `findBore()` with bundle
- `internal/qr`: compact `GenerateASCII` and `GeneratePNG`
- `internal/app`: `ListFiles` (sorted, filters dirs/dotfiles) and `ResolveFile`

Stack: Go 1.23+, `bubbletea` + `bubbles` + `lipgloss`, `skip2/go-qrcode`, stdlib `net/http`, `embed`, `os/exec`.

## Requirements

- Go 1.23+
- `bore` **or** `ssh` (for `serveo`/`lhr`) **or** `cloudflared` in `PATH` — but `envia` already includes `bore` bundle in `third-party/bore/` and `install.sh`/`install.ps1` download the bundle automatically, so normally no manual install needed.
- Tested on **macOS M1** and **Linux** (amd64/arm64, `CGO_ENABLED=0`). Windows with `bore.exe` bundle coming soon.

## Roadmap

**v0.1.0:** 1 file, preview, download, Range, Cloudflare+Bore

**v0.2.0:** bore as instant default (bundle MIT), fallback `serveo`/`localhost.run` via `ssh`, link outside box

**v0.2.1:** QR code in terminal and page (`/qr`)

**v0.2.3 (current):** compact QR (~13 lines), `bore` bundle and `serveo`/`lhr` fast

**v0.3:** multiple files, ZIP auto, multiple selection

**Future:** optional rendezvous server, password, expiration, P2P, extra providers

## Structure

```
envia/
├── cmd/envia/main.go, web.go, web/*
├── internal/app/*, httpserver/*, tunnel/*, qr/*
├── third-party/bore/ (bundle MIT)
├── scripts/test_ssh_fallbacks.py
├── docs/ (this site)
├── install.sh / install.ps1
├── Makefile
└── go.mod
```
