# envia v0.2.3

> **Compartilhe um arquivo sem nuvem — seu PC vira servidor temporário via bore. Um comando, um link.**

[![Go Version](https://img.shields.io/badge/go-1.23+-00ADD8?logo=go)](https://go.dev)
[![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)](https://github.com/Saimonsanbr/envia)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

```bash
envia video.mp4
# → https://xxxxx.trycloudflare.com
```

Quem recebe só abre o link no navegador — sem cadastro, sem app, sem upload pra nuvem do `envia`.

---

## Aviso — projeto entusiasta em desenvolvimento

O `envia` é um projeto **entusiasta, feito no tempo livre**. Ele funciona, mas ainda pode ter bugs, arestas e comportamentos inesperados — principalmente fora do ambiente onde foi testado.

- **Testado apenas em macOS com Apple Silicon (M1)** até agora.
- **Não testado ainda em Linux e Windows** — builds para essas plataformas estão no roadmap e devem sair nas próximas releases. Se você testar, conta pra gente como foi!
- A ideia é manter tudo **simples, rápido e sem cadastro**, mas a v0.2.0 ainda é um MVP de um único arquivo por vez.

Se encontrar algo estranho, abre uma issue. Toda ajuda é bem-vinda — e obrigado por testar tão cedo!

---

## Sumário

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

> `envia` fica disponível em qualquer pasta depois de instalado — basta ter `~/.local/bin` no `PATH`. O `cloudflared` precisa estar instalado separado (veja abaixo).

### Via curl (recomendado — macOS/Linux)

Instala o binário certo para seu sistema em `~/.local/bin/envia` (já com `bore` bundle):

```bash
curl -fsSL https://raw.githubusercontent.com/Saimonsanbr/envia/main/install.sh | sh
# com versão específica:
curl -fsSL https://raw.githubusercontent.com/Saimonsanbr/envia/main/install.sh | sh -s -- v0.2.3
# escolher pasta:
INSTALL_DIR=/usr/local/bin curl -fsSL https://raw.githubusercontent.com/Saimonsanbr/envia/main/install.sh | sh
```

Se `~/.local/bin` não estiver no `PATH`, adicione:

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc && source ~/.zshrc
# ou ~/.bashrc
```

### Via PowerShell (Windows — sem admin)

> **Segurança:** este script é 100% auditável e open source (`https://github.com/Saimonsanbr/envia/blob/main/install.ps1`). Ele só baixa `envia.exe` e `bore.exe` das releases oficiais e move para `%LOCALAPPDATA%\Programs\envia`. Não confie em scripts de qualquer pessoa — sempre verifique a URL e o código antes de executar.

Instala em `%LOCALAPPDATA%\Programs\envia` (sem precisar de admin) e adiciona ao `PATH` do usuário:

```powershell
# Opção 1 — direto (recomendado)
iwr -useb https://raw.githubusercontent.com/Saimonsanbr/envia/main/install.ps1 | iex

# Opção 2 — baixar, inspecionar e depois executar (mais seguro)
iwr https://raw.githubusercontent.com/Saimonsanbr/envia/main/install.ps1 -OutFile install.ps1
# abra install.ps1 no editor e verifique, depois:
powershell -ExecutionPolicy Bypass -File install.ps1
# ou com versão específica
powershell -ExecutionPolicy Bypass -File install.ps1 -Version v0.2.3
```

Se o Windows bloquear por `ExecutionPolicy`:

```powershell
# Ver atual
Get-ExecutionPolicy -Scope CurrentUser
# Permitir scripts do seu usuário (sem admin, reversível)
Set-ExecutionPolicy -Scope CurrentUser -ExecutionPolicy RemoteSigned -Force
# Ou só para esta execução
powershell -ExecutionPolicy Bypass -File install.ps1
```

O instalador já baixa o `bore.exe` bundle (`v0.6.0`, MIT) para a mesma pasta do `envia.exe`, então **não precisa `cargo install`** no Windows. Se o bundle falhar, o `envia` mostra o tutorial por OS.

### Via go install

```bash
go install github.com/Saimonsanbr/envia/cmd/envia@latest
# binário vai para $(go env GOPATH)/bin/envia
```

### Via download direto (sem script)

```bash
# Linux amd64
curl -L https://github.com/Saimonsanbr/envia/releases/latest/download/envia-linux-amd64 -o envia && chmod +x envia && sudo mv envia /usr/local/bin/
# Linux arm64
curl -L https://github.com/Saimonsanbr/envia/releases/latest/download/envia-linux-arm64 -o envia && chmod +x envia && sudo mv envia /usr/local/bin/
# macOS arm64 (M1/M2)
curl -L https://github.com/Saimonsanbr/envia/releases/latest/download/envia-darwin-arm64 -o envia && chmod +x envia && sudo mv envia /usr/local/bin/
# macOS amd64 (Intel)
curl -L https://github.com/Saimonsanbr/envia/releases/latest/download/envia-darwin-amd64 -o envia && chmod +x envia && sudo mv envia /usr/local/bin/
# Windows (PowerShell)
# Baixe https://github.com/Saimonsanbr/envia/releases/latest/download/envia-windows-amd64.exe
```

### Compilar do código

```bash
git clone https://github.com/Saimonsanbr/envia.git
cd envia
go build -trimpath -ldflags "-s -w" -o envia ./cmd/envia
# ou
make build  # gera bin/envia
```

### Provider de túnel

O `envia` tenta ser **portable**: por padrão ele já inclui o binário `bore` pré-compilado (MIT, `third-party/bore/LICENSE`), então **normalmente você não precisa instalar nada manualmente**.

Se o binário bundle falhar por algum motivo, o `envia` mostra:

```
bore não encontrado. Instale:
  macOS: brew install bore  ou  cargo install bore-cli
  Linux: cargo install bore-cli
  Windows: cargo install bore-cli  ou  baixe bore.exe em https://github.com/ekzhang/bore/releases
  Tutorial: https://github.com/Saimonsanbr/envia#instalação
```

**Por que `bore` e `ssh` em vez de `cloudflared`?**
Testamos `cloudflared` exaustivamente e deu vários problemas: `429 Too Many Requests` em rajadas, `DNS_PROBE_POSSIBLE` por propagação lenta de `*.trycloudflare.com` e health de `6s` que estourava. O `bore` (`bore.pub`) é instantâneo (2-4s), leve e sem conta; `serveo.net` e `localhost.run` via `ssh` são fallbacks `ssh -R` igualmente simples e sem rate limit agressivo. Por isso `auto` agora é `bore → serveo → localhost.run`, e `cloudflared` ficou só para `--provider cloudflare` (avançado, ex: quem tem domínio próprio e quer configurar `provider=cloudflare` no `config.json`).

**Instalação manual por OS (só se o bundle falhar):**

**macOS:**
```bash
brew install bore
# ou
cargo install bore-cli
bore --version
```

**Linux:**
```bash
cargo install bore-cli
# precisa Rust: https://rustup.rs
bore --version
```

**Windows:**
```powershell
cargo install bore-cli
# ou baixe bore.exe em https://github.com/ekzhang/bore/releases/download/v0.6.0/bore-v0.6.0-x86_64-pc-windows-msvc.zip
# descompacte e coloque bore.exe no PATH ou ao lado de envia.exe
bore --version
```

O modo `--provider auto` (padrão) tenta `bore` 3× (sem health, instantâneo) → `serveo` → `localhost.run`. Se `bore` não estiver instalado *e* o bundle falhar, tenta `cloudflare` como último recurso. Use `--provider cloudflare`, `--provider serveo` ou `--provider localhost.run` explicitamente se quiser.

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
4. Chama `bore local PORT --to bore.pub` (padrão, instantâneo, 2-4s) ou `cloudflared tunnel --url http://127.0.0.1:PORTA` se `--provider cloudflare`.
5. Lê a saída do processo e captura a URL (`bore.pub:PORT` ou `https://*.trycloudflare.com`).
6. Para `bore`, retorna imediatamente (sem health); para `cloudflare`, valida com `GET /` rápido (retry de até 3 tentativas, 2.5s cada, com `LookupIP` fail-fast). Se `bore` falhar, tenta `cloudflare` como último recurso quando em `auto`.
7. Mostra o link com spinner `bubbles/spinner` durante a criação.
8. Serve o arquivo com streaming e fica esperando até `Ctrl+C`.

Enquanto o spinner gira, você só vê “Criando link…” — o retry de DNS (`DNS_PROBE_POSSIBLE`) acontece por baixo sem poluir o terminal.

---

## Como funciona (por dentro)

> Segunda etapa, mais técnica mas ainda legível. Depois vamos expandir em `docs/`.

**HTTP (`net/http` puro, sem frameworks):**
- `127.0.0.1:0` → porta aleatória, nunca `0.0.0.0` na 0.2.0.
- Endpoints: `GET /` (HTML), `GET /preview` (inline), `GET /download` (attachment), `GET /preview.css`, `GET /app.js`.
- `http.ServeContent` com `os.Open` → não carrega 20GB na RAM, suporta `Range: bytes=...` (`206 Partial Content`) para seek e retomada.
- Headers: `X-Content-Type-Options: nosniff`, `Referrer-Policy: no-referrer`, `Accept-Ranges: bytes`.
- Só serve o arquivo escolhido — sem `../` ou navegação.

**Página (`embed.FS`):**
- `cmd/envia/web/*` (`index.html`, `preview.css`, `app.js`) embutido no binário via `//go:embed web/*`.
- `html/template` com escaping para nome com espaços/acentos/unicode.
- Preview: `image` → `<img>`, `video` → `<video controls>`, `audio` → `<audio>`, `pdf` → `<iframe>`, fallback genérico. CSS puro com `prefers-color-scheme` (claro/escuro), 100% offline (sem CDN).

**Túnel (`os/exec`):**
- `exec.LookPath` verifica `bore`/`cloudflared` no `PATH` (`bore` é padrão agora, `cloudflared` só se `--provider cloudflare`).
- `exec.CommandContext` com `signal.NotifyContext` para matar filhos no `Ctrl+C`.
- Regex `bore\.pub:(\d+)` e `https://[a-zA-Z0-9-]+\.trycloudflare\.com` com strip de ANSI.
- `bore`: `12s` para URL, sem health (instantâneo, 3 retries com `100ms` gap). `cloudflare`: `15s` para URL + health `2.5s` (`LookupIP` fail-fast + `GET` com `Timeout 1.5s`) por tentativa, 3 tentativas.

**UI (`charmbracelet/bubbletea` + `bubbles`):**
- Modelo com estados `picking → loading → ready`. `list` para arquivos, `spinner.Dot` para carregando.
- Fallback para modo plain (`app.Run`) quando não há TTY (ex: pipes/CI).

---

## Página e preview

A página é minimalista, responsiva e funciona no celular:

- Imagens (`.jpg`, `.png`, `.webp` etc) com `<img>`
- Vídeos (`.mp4`, `.webm`) com `<video controls>`
- Áudios (`.mp3`, `.wav`) com `<audio controls>`
- PDFs com `<iframe>`, outros com ícone de arquivo + botão **Baixar arquivo**

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

Mas o tráfego **passa pelo provider de túnel** (Bore/Serveo) — ele pode ver o tráfego. O link é temporário e sem autenticação na v0.2.0: quem tiver a URL acessa o arquivo enquanto você deixar o `envia` rodando.

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

**v0.1.0:** 1 arquivo, preview, download, Range, Cloudflare+Bore, UI spinner, retry de DNS.

**v0.2.0:** bore como padrão instantâneo (bundle MIT), fallback serveo/localhost.run via ssh, link fora da box.

**v0.2.1:** QR code no terminal e na página (`/qr`).

**v0.2.3 (atual):** QR compacto (~13 linhas) para não esconder o link em terminal estreito, `bore` bundle e `serveo`/`lhr` rápidos.

**v0.3:** múltiplos arquivos, ZIP automático, seleção múltipla.

**Futuro:** rendezvous server opcional, senha, expiração, P2P, providers extras.

---

## Licença

MIT — faça o que quiser, só mantenha o aviso.
