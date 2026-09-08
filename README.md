# envia v0.1.0

`envia` transforma temporariamente seu computador em um servidor de arquivos acessível pela internet.

Com um único comando você compartilha **um arquivo** sem fazer upload para nuvem própria:

```bash
envia video.mp4
# → https://xxxxx.trycloudflare.com
```

Quem recebe só precisa abrir o link no navegador — sem cadastro, sem app.

## O que é

- O arquivo **permanece no seu computador**
- Seu computador vira servidor HTTP local (`127.0.0.1:PORTA` aleatória)
- Um túnel público (Cloudflare Tunnel) expõe temporariamente esse servidor
- O link deixa de funcionar quando você pressiona `Ctrl+C`

```
arquivo → servidor local → túnel → navegador
  (disco)    (127.0.0.1)   (cloudflared/bore)  (download/preview)
```

> O `envia` não possui backend próprio de armazenamento. Não há upload para servidor do `envia`.

## Instalação

### 1. Instalar o binário

```bash
go build -trimpath -ldflags "-s -w" -o envia ./cmd/envia
# ou
make build  # gera bin/envia
```

Ou baixe o binário da release (quando disponível).

### 2. Instalar provider de túnel

**Cloudflare (recomendado, padrão):**

```bash
brew install cloudflared
# ou veja https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/
cloudflared --version
```

**Bore (fallback opcional):**

```bash
cargo install bore-cli
# ou
brew install bore
bore --version
```

O `envia` usa `cloudflared` por padrão. Se não estiver instalado e `bore` estiver, usa `bore` automaticamente no modo `--provider auto`.

## Uso

```bash
# Selecionar arquivo interativamente (lista arquivos da pasta atual)
envia

# Enviar arquivo direto
envia video.mp4
envia ./foto.jpg
envia /Users/saimo/Videos/video.mp4

# Escolher provider
envia --provider cloudflare video.mp4
envia --provider bore video.mp4
envia --provider auto video.mp4  # padrão: tenta cloudflare, depois bore

# Versão
envia --version
envia -v
```

Interface sem argumento abre seletor com `bubbletea`/`bubbles` list: `↑↓` navega, digite para filtrar, `Enter` seleciona, `Esc`/`Ctrl+C` cancela.

Após escolher o arquivo:

```
envia v0.1.0

Compartilhe um arquivo diretamente do seu computador.

Arquivo
  video.mp4
  327.4 MB

Servidor
  http://127.0.0.1:43821

Túnel
  Cloudflare

Link público
  https://abc123.trycloudflare.com

O arquivo continua no seu computador.
Nenhum upload foi feito para o envia.

Aguardando downloads...

Pressione Ctrl+C para encerrar.
```

Pressione `Ctrl+C` para encerrar servidor e túnel — o link para de funcionar imediatamente.

## Funcionamento

1. Valida arquivo (não pode ser diretório)
2. Inicia servidor HTTP em `127.0.0.1:PORTA` aleatória (`net/http`)
3. Inicia `cloudflared tunnel --url http://127.0.0.1:PORTA` (ou `bore local PORT --to bore.pub`)
4. Captura URL pública da saída do processo (regex `https://*.trycloudflare.com` e `bore.pub:PORT`)
5. Serve a página em `/` e o arquivo em `/preview` (inline) e `/download` (attachment)
6. Usa `http.ServeContent` para streaming sem carregar arquivo na RAM e suporte a `Range: bytes=...` (`206 Partial Content`) para seek e retomada
7. Aguarda até `SIGINT`/`SIGTERM`, faz shutdown limpo

### Endpoints

- `GET /` → página HTML com preview e botão baixar
- `GET /preview` → arquivo com `Content-Type` correto (inline)
- `GET /download` → mesmo arquivo com `Content-Disposition: attachment`
- `GET /preview.css` e `GET /app.js` → assets embutidos

A página é `HTML/CSS/JS` puro, sem CDN, embutida via `embed.FS`. Detecta preview: `image` → `<img>`, `video` → `<video controls>`, `audio` → `<audio>`, fallback genérico.

### Headers

- `X-Content-Type-Options: nosniff`
- `Referrer-Policy: no-referrer`
- `Accept-Ranges: bytes`

Servidor serve **somente** o arquivo selecionado — sem navegação no filesystem.

## Privacidade

> O `envia` não armazena seu arquivo. Nenhum upload é feito para backend próprio.

O tráfego, entretanto, passa pelo provider de túnel que você escolher (Cloudflare ou Bore). O provedor pode ver o tráfego que passa pelo túnel.

O link é temporário: quem tiver a URL pode acessar o arquivo enquanto o `envia` estiver rodando. Não há autenticação na v0.1.0.

## Desenvolvimento

```bash
make run      # go run ./cmd/envia
make build    # cria bin/envia
make test     # go test ./...
make fmt      # gofmt -w .
make clean    # remove bin/
```

Testes cobrem:

- HTTP server: página, preview, download, Content-Type, Range 206, headers
- Tunnel: parsing de URL sem depender de rede
- File picker: ListFiles e ResolveFile

## Requisitos

- Go 1.23+
- `cloudflared` ou `bore` no `PATH`
- macOS / Linux / Windows (amd64/arm64)

## Licença

MIT
