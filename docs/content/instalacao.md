# Instalação

`envia` fica disponível em qualquer pasta depois de instalado — basta ter `~/.local/bin` no `PATH`.

## Via curl (recomendado)

**macOS / Linux:**

```bash
curl -fsSL https://raw.githubusercontent.com/Saimonsanbr/envia/main/install.sh | sh
# com versão específica
curl -fsSL https://raw.githubusercontent.com/Saimonsanbr/envia/main/install.sh | sh -s -- v0.2.3
# escolher pasta
INSTALL_DIR=/usr/local/bin curl -fsSL https://raw.githubusercontent.com/Saimonsanbr/envia/main/install.sh | sh
```

Se `~/.local/bin` não estiver no `PATH`:

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc && source ~/.zshrc
# ou ~/.bashrc
```

**Windows (PowerShell, sem admin):**

```powershell
iwr -useb https://raw.githubusercontent.com/Saimonsanbr/envia/main/install.ps1 | iex

# ou baixar, inspecionar e depois executar (recomendado)
iwr https://raw.githubusercontent.com/Saimonsanbr/envia/main/install.ps1 -OutFile install.ps1
# abra o install.ps1 no editor para auditar, depois:
powershell -ExecutionPolicy Bypass -File install.ps1
```

> **Segurança:** O `install.ps1` é 100% auditável em https://github.com/Saimonsanbr/envia/blob/main/install.ps1 — ele só baixa `envia.exe` e `bore.exe` das releases oficiais. Não confie em scripts de qualquer pessoa, sempre verifique a URL.

O instalador Windows já baixa o `bore.exe` bundle (`v0.6.0`, MIT) para a mesma pasta do `envia.exe`, então **não precisa `cargo install`** no Windows.

## Via go install

```bash
go install github.com/Saimonsanbr/envia/cmd/envia@latest
# binário vai para $(go env GOPATH)/bin/envia
```

## Via download direto

```bash
# Linux amd64
curl -L https://github.com/Saimonsanbr/envia/releases/latest/download/envia-linux-amd64 -o envia && chmod +x envia && sudo mv envia /usr/local/bin/
# Linux arm64
curl -L https://github.com/Saimonsanbr/envia/releases/latest/download/envia-linux-arm64 -o envia && chmod +x envia && sudo mv envia /usr/local/bin/
# macOS arm64 (M1/M2)
curl -L https://github.com/Saimonsanbr/envia/releases/latest/download/envia-darwin-arm64 -o envia && chmod +x envia && sudo mv envia /usr/local/bin/
# macOS amd64 (Intel)
curl -L https://github.com/Saimonsanbr/envia/releases/latest/download/envia-darwin-amd64 -o envia && chmod +x envia && sudo mv envia /usr/local/bin/
# Windows PowerShell
# Baixe https://github.com/Saimonsanbr/envia/releases/latest/download/envia-windows-amd64.exe
```

## Compilar do código

```bash
git clone https://github.com/Saimonsanbr/envia.git
cd envia
go build -trimpath -ldflags "-s -w" -o envia ./cmd/envia
# ou
make build  # gera bin/envia
```

## Provider de túnel

Por padrão o `envia` já inclui o `bore` pré-compilado (`third-party/bore/LICENSE`, MIT), então normalmente **não precisa instalar nada manualmente**.

Se o bundle falhar, o `envia` mostra:

```
bore não encontrado. Instale:
  macOS: brew install bore  ou  cargo install bore-cli
  Linux: cargo install bore-cli
  Windows: cargo install bore-cli  ou  baixe bore.exe em https://github.com/ekzhang/bore/releases
  Tutorial: https://github.com/Saimonsanbr/envia#instalação
```

**Por que `bore` e `ssh` em vez de `cloudflared`?**
Testamos `cloudflared` exaustivamente e deu `429 Too Many Requests` em rajadas e `DNS_PROBE_POSSIBLE` por propagação lenta de `*.trycloudflare.com`. O `bore` (`bore.pub`) é instantâneo (2-4s) e leve; `serveo.net` e `localhost.run` via `ssh -R` são fallbacks igualmente simples. Por isso `auto` agora é `bore → serveo → localhost.run`, e `cloudflared` ficou só para `--provider cloudflare` (avançado, ex: quem tem domínio próprio e configura `provider=cloudflare` no `config.json`).

**Manual por OS (só se o bundle falhar):**

*macOS:*
```bash
brew install bore
# ou
cargo install bore-cli
```

*Linux:*
```bash
cargo install bore-cli
# precisa Rust: https://rustup.rs
```

*Windows:*
```powershell
cargo install bore-cli
# ou baixe bore.exe em https://github.com/ekzhang/bore/releases/download/v0.6.0/bore-v0.6.0-x86_64-pc-windows-msvc.zip
```
