# Como funciona

## Visão geral

1. Você escolhe um arquivo (argumento ou seletor).
2. O `envia` valida se é arquivo (não diretório) e pega tamanho.
3. Abre um HTTP server só em `127.0.0.1` com porta que o sistema escolhe.
4. Chama `bore local PORT --to bore.pub` (padrão, 2-4s) ou `cloudflared tunnel --url http://127.0.0.1:PORTA` se `--provider cloudflare`.
5. Lê a saída do processo e captura a URL (`bore.pub:PORT` ou `https://*.trycloudflare.com`).
6. Para `bore`, retorna imediatamente (sem health); para `cloudflare`, valida com `GET /` rápido (retry de até 3 tentativas, 2.5s cada). Se `bore` falhar 3×, tenta `serveo` (`ssh -R 80:localhost:PORT serveo.net` → `https://*.serveo.net`) e depois `localhost.run` (`https://*.lhr.life`).
7. Mostra o link com spinner `bubbles/spinner` e QR code (`skip2/go-qrcode`, <5ms) durante a criação.
8. Serve o arquivo com streaming e fica esperando até `Ctrl+C`.

Enquanto o spinner gira, você só vê “Criando link…” — o retry acontece por baixo.

## Por dentro

**HTTP (`net/http` puro, sem frameworks):**
- `127.0.0.1:0` → porta aleatória, nunca `0.0.0.0` na 0.2.3.
- Endpoints: `GET /` (HTML), `GET /preview` (inline), `GET /download` (attachment), `GET /preview.css`, `GET /app.js`, `GET /qr` (PNG do QR).
- `http.ServeContent` com `os.Open` → não carrega 20GB na RAM, suporta `Range: bytes=...` (`206 Partial Content`) para seek e retomada.
- Headers: `X-Content-Type-Options: nosniff`, `Referrer-Policy: no-referrer`, `Accept-Ranges: bytes`.
- Só serve o arquivo escolhido — sem `../` ou navegação.

**Página (`embed.FS`):**
- `cmd/envia/web/*` (`index.html`, `preview.css`, `app.js`) embutido no binário via `//go:embed web/*`.
- `html/template` com escaping para nome com espaços/acentos/unicode.
- Preview: `image` → `<img>`, `video` → `<video controls>`, `audio` → `<audio>`, `pdf` → `<iframe>`, fallback genérico. CSS puro com `prefers-color-scheme`, 100% offline (sem CDN).

**Túnel (`os/exec`):**
- `exec.LookPath` verifica `bore`/`cloudflared`/`ssh` no `PATH` (`bore` é padrão, bundle em `third-party/bore/` ou ao lado do `envia`).
- `exec.CommandContext` com `signal.NotifyContext` para matar filhos no `Ctrl+C`.
- Regex `bore\.pub:(\d+)`, `https://[a-zA-Z0-9-]+\.serveo\.net`, `https://[a-zA-Z0-9-]+\.lhr\.life` e `https://[a-zA-Z0-9-]+\.trycloudflare\.com` com strip de ANSI.
- `bore`: `12s` para URL, sem health (instantâneo, 3 retries com `100ms` gap). `cloudflare`: `15s` para URL + health `2.5s` por tentativa, 3 tentativas. `serveo`/`lhr`: `15s` via `ssh`.

**UI (`charmbracelet/bubbletea` + `bubbles`):**
- Modelo com estados `picking → loading → ready`. `list` para arquivos, `spinner.Dot` para carregando.
- QR code cacheado no `tunnelReadyMsg` (gerado uma vez, não a cada frame).
- Fallback para modo plain (`app.Run`) quando não há TTY (ex: pipes/CI).

**QR code:**
- `internal/qr/qr.go` com `skip2/go-qrcode` (Low recovery, 256x256 PNG para web, ASCII com half-blocks `▀▄█` para terminal, ~13 linhas vs 30 antes).

## Diagrama

```
Seu PC (envia)
  └─ arquivo no disco
  └─ http://127.0.0.1:PORT (ServeContent)
  └─ bore local PORT --to bore.pub  ──┐
                                      │ túnel
                                      ▼
                                   bore.pub:PORT
                                      │
                                      ▼
                                   Navegador / celular (QR)
                                    → GET /preview (Range)
                                    → GET /download (attachment)
                                    → GET /qr (PNG)
```
