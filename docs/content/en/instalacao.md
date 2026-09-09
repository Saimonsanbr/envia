# Installation

`envia` is available in any folder after installation — just have `~/.local/bin` in your `PATH`.

## Via curl (recommended — macOS/Linux)

Install the right binary for your system to `~/.local/bin/envia` (with `bore` bundle):

```bash
curl -fsSL https://raw.githubusercontent.com/Saimonsanbr/envia/main/install.sh | sh
# with specific version
curl -fsSL https://raw.githubusercontent.com/Saimonsanbr/envia/main/install.sh | sh -s -- v0.2.3
# choose folder
INSTALL_DIR=/usr/local/bin curl -fsSL https://raw.githubusercontent.com/Saimonsanbr/envia/main/install.sh | sh
```

If `~/.local/bin` is not in your `PATH`:

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc && source ~/.zshrc
# or ~/.bashrc
```

**Windows (PowerShell, no admin):**

```powershell
iwr -useb https://raw.githubusercontent.com/Saimonsanbr/envia/main/install.ps1 | iex

# or download, inspect, then run (recommended)
iwr https://raw.githubusercontent.com/Saimonsanbr/envia/main/install.ps1 -OutFile install.ps1
# open install.ps1 in editor to audit, then:
powershell -ExecutionPolicy Bypass -File install.ps1
```

> **Security:** This script is 100% auditable at https://github.com/Saimonsanbr/envia/blob/main/install.ps1 — it only downloads `envia.exe` and `bore.exe` from official releases. Don't trust scripts from anyone — always check the URL and code.

The Windows installer already downloads the `bore.exe` bundle (`v0.6.0`, MIT) to the same folder as `envia.exe`, so **no `cargo install` needed** on Windows.

## Via go install

```bash
go install github.com/Saimonsanbr/envia/cmd/envia@latest
# binary goes to $(go env GOPATH)/bin/envia
```

## Via direct download

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
# Download https://github.com/Saimonsanbr/envia/releases/latest/download/envia-windows-amd64.exe
```

## Build from source

```bash
git clone https://github.com/Saimonsanbr/envia.git
cd envia
go build -trimpath -ldflags "-s -w" -o envia ./cmd/envia
# or
make build  # generates bin/envia
```

## Tunnel provider

`envia` tries to be **portable**: by default it already includes the `bore` pre-compiled binary (MIT, `third-party/bore/LICENSE`), so **you normally don't need to install anything manually**.

If the bundle fails for some reason, `envia` shows:

```
bore not found. Install:
  macOS: brew install bore  or  cargo install bore-cli
  Linux: cargo install bore-cli
  Windows: cargo install bore-cli  or  download bore.exe at https://github.com/ekzhang/bore/releases
  Tutorial: https://github.com/Saimonsanbr/envia#installation
```

**Why `bore` and `ssh` instead of `cloudflared`?**
We tested `cloudflared` extensively and got `429 Too Many Requests` in bursts and `DNS_PROBE_POSSIBLE` due to slow propagation of `*.trycloudflare.com`. `bore` (`bore.pub`) is instant (2-4s) and lightweight; `serveo.net` and `localhost.run` via `ssh -R` are equally simple fallbacks. That's why `auto` is now `bore → serveo → localhost.run`, and `cloudflared` is only for `--provider cloudflare` (advanced, e.g., who has own domain and sets `provider=cloudflare` in `config.json`).

**Manual install per OS (only if bundle fails):**

*macOS:*
```bash
brew install bore
# or
cargo install bore-cli
```

*Linux:*
```bash
cargo install bore-cli
# needs Rust: https://rustup.rs
```

*Windows:*
```powershell
cargo install bore-cli
# or download bore.exe at https://github.com/ekzhang/bore/releases/download/v0.6.0/bore-v0.6.0-x86_64-pc-windows-msvc.zip
```
