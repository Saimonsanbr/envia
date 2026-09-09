# Desenvolvimento

```bash
make run      # go run ./cmd/envia
make build    # cria bin/envia
make test     # go test ./...
make fmt      # gofmt -w .
make clean    # remove bin/
```

Testes cobrem:

- `internal/httpserver`: página, `Content-Type`, `Content-Disposition: attachment`, `Range 206`, headers, arquivo grande via streaming, `/qr` PNG
- `internal/tunnel`: parsing de URL (`bore.pub`, `serveo.net`, `lhr.life`, `trycloudflare.com`) sem rede, `findBore()` com bundle
- `internal/qr`: `GenerateASCII` compacto e `GeneratePNG`
- `internal/app`: `ListFiles` (ordenado, filtra diretórios/dotfiles) e `ResolveFile`

Stack: Go 1.23+, `bubbletea` + `bubbles` + `lipgloss`, `skip2/go-qrcode`, stdlib `net/http`, `embed`, `os/exec`.

## Requisitos

- Go 1.23+
- `bore` **ou** `ssh` (para `serveo`/`lhr`) **ou** `cloudflared` no `PATH` — mas `envia` já inclui `bore` bundle em `third-party/bore/` e `install.sh`/`install.ps1` baixam o bundle automaticamente, então normalmente não precisa instalar manualmente.
- Testado em **macOS M1** e **Linux** (amd64/arm64, `CGO_ENABLED=0`). Windows será testado em breve com `bore.exe` bundle.

## Roadmap

**v0.1.0:** 1 arquivo, preview, download, Range, Cloudflare+Bore

**v0.2.0:** bore como padrão instantâneo (bundle MIT), fallback `serveo`/`lhr` via `ssh`, link fora da box

**v0.2.1:** QR code no terminal e na página (`/qr`)

**v0.2.3 (atual):** QR compacto (~13 linhas), `bore` bundle e `serveo`/`lhr` rápidos

**v0.3:** múltiplos arquivos, ZIP automático, seleção múltipla

**Futuro:** rendezvous server opcional, senha, expiração, P2P, providers extras

## Estrutura

```
envia/
├── cmd/envia/main.go, web.go, web/*
├── internal/app/*, httpserver/*, tunnel/*, qr/*
├── third-party/bore/ (bundle MIT)
├── scripts/test_ssh_fallbacks.py
├── docs/ (este site)
├── install.sh / install.ps1
├── Makefile
└── go.mod
```
