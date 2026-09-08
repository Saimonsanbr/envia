# envia v0.1.0

> **Compartilhe um arquivo sem nuvem — seu PC vira servidor temporário via Cloudflare Tunnel. Um comando, um link.**

[![Go Version](https://img.shields.io/badge/go-1.23+-00ADD8?logo=go)](https://go.dev)
[![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)](https://github.com/Saimonsanbr/envia)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

```bash
envia video.mp4
# → https://xxxxx.trycloudflare.com
```

Quem recebe só abre o link no navegador — sem cadastro, sem app, sem upload pra nuvem do `envia`.

---

## ⚠️ Aviso — projeto entusiasta em desenvolvimento

O `envia` é um projeto **entusiasta, feito no tempo livre**. Ele funciona, mas ainda pode ter bugs, arestas e comportamentos inesperados — principalmente fora do ambiente onde foi testado.

- **Testado apenas em macOS com Apple Silicon (M1)** até agora.
- **Não testado ainda em Linux e Windows** — builds para essas plataformas estão no roadmap e devem sair nas próximas releases. Se você testar, conta pra gente como foi!
- A ideia é manter tudo **simples, rápido e sem cadastro**, mas a v0.1.0 ainda é um MVP de um único arquivo por vez.

Se encontrar algo estranho, abre uma issue. Toda ajuda é bem-vinda — e obrigado por testar tão cedo!

---

## 📚 Sumário

- [O que é](#o-que-é)
- [Instalação](#instalação)
- [Uso](#uso)
- [Como funciona (visão geral)](#como-funciona-visão-geral)
- [Como funciona (por dentro)](#como-funciona-por-dentro)
- [Página e preview](#página-e-preview)
- [Configuração](#configuração)
- [Privacidade](#privacidade)
- [Desenvolvimento](#desenvolvimento)
- [Requisitos](#requisitos)
- [Roadmap](#roadmap)
- [Licença](#licença)

---

## O que é

`envia` transforma temporariamente seu computador em um servidor de arquivos.

- O arquivo **continua no seu disco** — não faz upload pra nuvem do `envia`.
- Seu PC abre um servidor HTTP local em `127.0.0.1` com porta aleatória.
- Um túnel público (`cloudflared` ou `bore`) expõe esse servidor na internet.
- Você compartilha a URL, a pessoa baixa direto de você. Quando você dá `Ctrl+C`, o link morre.

```
arquivo → servidor local → túnel → navegador
 (disco)    (127.0.0.1)   (cloudflared/bore)  (preview/download)
```

> Sem backend próprio, sem conta, sem painel. Só um binário.

---

## Instalação

### 1. Binário `envia`

Por enquanto, compile localmente (releases pra Linux/Windows vêm em breve):

```bash
go build -trimpath -ldflags "-s -w" -o envia ./cmd/envia
# ou
make build  # gera bin/envia
```

> Em breve: `curl -L https://github.com/Saimonsanbr/envia/releases/latest/download/envia-darwin-arm64` etc.

### 2. Provider de túnel

**Cloudflare (recomendado, padrão):**

```bash
brew install cloudflared
# https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/
cloudflared --version
```

**Bore (fallback):**

```bash
cargo install bore-cli
# ou
brew install bore
bore --version
```

O modo `--provider auto` (padrão) tenta `cloudflared` primeiro e só usa `bore` se precisar.

---

## Uso

```bash
# interativo — lista arquivos da pasta atual
envia

# direto
envia video.mp4
envia ./foto.jpg
envia /Users/saimo/Videos/video.mp4

# escolher provider
envia --provider cloudflare video.mp4
envia --provider bore video.mp4
envia --provider auto video.mp4  # padrão

# configs
envia --config              # mostra config atual
envia --config ocultos      # alterna exibição de arquivos ocultos (.files)

# versão
envia --version
envia -v
```

**Seletor interativo** (`bubbletea`/`bubbles`): `↑↓` navega, digite para filtrar, `Enter` seleciona, `Esc` cancela. Arquivos ocultos (`.DS_Store`, `.env`) ficam escondidos por padrão — use `--config ocultos` para mostrar.

Depois de escolher:

```
╭──────────────────────────────────────────╮
│ Arquivo                                  │
│   video.mp4  327.4 MB                     │
│ Servidor                                 │
│   http://127.0.0.1:43821                  │
│ Túnel                                    │
│   cloudflare                             │
│ Link público                             │
│   https://abc123.trycloudflare.com       │
╰──────────────────────────────────────────╯
O arquivo continua no seu computador.
Aguardando downloads... • Ctrl+C para encerrar
```

`Ctrl+C` encerra servidor e túnel — o link para de funcionar na hora.

---

## Como funciona (visão geral)

1. Você escolhe um arquivo (argumento ou seletor).
2. O `envia` valida se é arquivo (não diretório) e pega tamanho.
3. Abre um HTTP server só em `127.0.0.1` com porta que o sistema escolhe.
4. Chama `cloudflared tunnel --url http://127.0.0.1:PORTA` (ou `bore local PORT --to bore.pub`).
5. Lê a saída do processo e captura a URL (`https://*.trycloudflare.com` ou `bore.pub:PORT`).
6. Valida a URL com um `GET /` rápido (retry transparente de até 5 tentativas, poucos segundos cada). Se Cloudflare falhar, tenta novo túnel; se 5 falharem, cai pro `bore`.
7. Mostra o link com spinner `bubbles/spinner` durante a criação.
8. Serve o arquivo com streaming e fica esperando até `Ctrl+C`.

Enquanto o spinner gira, você só vê “Criando link…” — o retry de DNS (`DNS_PROBE_POSSIBLE`) acontece por baixo sem poluir o terminal.

---

## Como funciona (por dentro)

> Segunda etapa, mais técnica mas ainda legível. Depois vamos expandir em `docs/`.

**HTTP (`net/http` puro, sem frameworks):**
- `127.0.0.1:0` → porta aleatória, nunca `0.0.0.0` na 0.1.0.
- Endpoints: `GET /` (HTML), `GET /preview` (inline), `GET /download` (attachment), `GET /preview.css`, `GET /app.js`.
- `http.ServeContent` com `os.Open` → não carrega 20GB na RAM, suporta `Range: bytes=...` (`206 Partial Content`) para seek e retomada.
- Headers: `X-Content-Type-Options: nosniff`, `Referrer-Policy: no-referrer`, `Accept-Ranges: bytes`.
- Só serve o arquivo escolhido — sem `../` ou navegação.

**Página (`embed.FS`):**
- `cmd/envia/web/*` (`index.html`, `preview.css`, `app.js`) embutido no binário via `//go:embed web/*`.
- `html/template` com escaping para nome com espaços/acentos/unicode.
- Preview: `image` → `<img>`, `video` → `<video controls>`, `audio` → `<audio>`, `pdf` → `<iframe>`, fallback genérico. CSS puro com `prefers-color-scheme` (claro/escuro), 100% offline (sem CDN).

**Túnel (`os/exec`):**
- `exec.LookPath` verifica `cloudflared`/`bore` no `PATH`.
- `exec.CommandContext` com `signal.NotifyContext` para matar filhos no `Ctrl+C`.
- Regex `https://[a-zA-Z0-9-]+\.trycloudflare\.com` e `bore\.pub:(\d+)` com strip de ANSI.
- Timeout 20s para URL + health 4s (`GET /` com `http.Client{Timeout:3s}`) por tentativa.

**UI (`charmbracelet/bubbletea` + `bubbles`):**
- Modelo com estados `picking → loading → ready`. `list` para arquivos, `spinner.Dot` para carregando.
- Fallback para modo plain (`app.Run`) quando não há TTY (ex: pipes/CI).

---

## Página e preview

A página é minimalista, responsiva e funciona no celular:

- Imagens (`.jpg`, `.png`, `.webp` etc) com `<img>`
- Vídeos (`.mp4`, `.webm`) com `<video controls>`
- Áudios (`.mp3`, `.wav`) com `<audio controls>`
- PDFs com `<iframe>`, outros com ícone `📄` + botão **Baixar arquivo**

Tudo vem do binário — quem recebe não instala nada.

---

## Configuração

Config salvo em `~/Library/Application Support/envia/config.json` (macOS) ou `~/.config/envia/config.json` (Linux):

```json
{
  "show_hidden": false
}
```

```bash
envia --config              # ver
envia --config ocultos      # toggle (esconde ↔ mostra .files)
```

Futuras flags (`--config` expansível) vão usar o mesmo arquivo.

---

## Privacidade

> O `envia` **não armazena** seu arquivo. Não tem upload pra backend próprio.

Mas o tráfego **passa pelo provider de túnel** (Cloudflare ou Bore) — ele pode ver o tráfego. O link é temporário e sem autenticação na v0.1.0: quem tiver a URL acessa o arquivo enquanto você deixar o `envia` rodando.

Não coletamos nada: sem analytics, sem trackers, sem conta, sem histórico. A ferramenta é stateless.

---

## Desenvolvimento

```bash
make run      # go run ./cmd/envia
make build    # cria bin/envia
make test     # go test ./...
make fmt      # gofmt -w .
make clean    # remove bin/
```

Testes:
- `internal/httpserver`: página, `Content-Type`, `Content-Disposition: attachment`, `Range 206`, headers, arquivo grande via streaming
- `internal/tunnel`: parsing de URL (sem rede)
- `internal/app`: `ListFiles` (ordenado, filtra diretórios/dotfiles) e `ResolveFile`

Stack: Go 1.23+, `bubbletea` + `bubbles` + `lipgloss`, stdlib `net/http`, `embed`, `os/exec`.

---

## Requisitos

- Go 1.23+
- `cloudflared` **ou** `bore` no `PATH`
- Testado em **macOS M1**; Linux/Windows em breve (arm64/amd64, `CGO_ENABLED=0`)

---

## Roadmap

**v0.1.0 (atual):** 1 arquivo, preview, download, Range, Cloudflare+Bore, UI spinner, retry de DNS.

**v0.2:** múltiplos arquivos, ZIP automático, QR code, seleção múltipla.

**Futuro:** rendezvous server opcional, senha, expiração, P2P, providers extras.

Releases para Linux/Windows virão via GitHub Actions assim que validarmos os builds.

---

## Licença

MIT — faça o que quiser, só mantenha o aviso.
