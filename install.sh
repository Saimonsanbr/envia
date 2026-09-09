#!/bin/sh
# envia installer - https://github.com/Saimonsanbr/envia
# Uso: curl -fsSL https://raw.githubusercontent.com/Saimonsanbr/envia/main/install.sh | sh
#      ou: curl -fsSL https://raw.githubusercontent.com/Saimonsanbr/envia/main/install.sh | sh -s -- v0.1.0
set -e

REPO="Saimonsanbr/envia"
BINARY="envia"
VERSION="${1:-latest}"
INSTALL_DIR="${INSTALL_DIR:-}"

# Cores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

info() { printf "${GREEN}==>${NC} %s\n" "$1"; }
warn() { printf "${YELLOW}⚠${NC} %s\n" "$1"; }
err() { printf "${RED}✗${NC} %s\n" "$1" >&2; }

# Detecta OS e ARCH
detect_platform() {
  OS="$(uname -s)"
  ARCH="$(uname -m)"
  case "$OS" in
    Linux) GOOS="linux" ;;
    Darwin) GOOS="darwin" ;;
    *) err "OS não suportado: $OS (use download manual da release)"; exit 1 ;;
  esac
  case "$ARCH" in
    x86_64|amd64) GOARCH="amd64" ;;
    arm64|aarch64) GOARCH="arm64" ;;
    *) err "Arch não suportado: $ARCH"; exit 1 ;;
  esac
  # Sufixo .exe só Windows (não coberto por este script)
  echo "${GOOS}-${GOARCH}"
}

# Escolhe diretório de instalação
choose_install_dir() {
  if [ -n "$INSTALL_DIR" ]; then
    echo "$INSTALL_DIR"
    return
  fi
  # Prefere ~/.local/bin se existir ou /usr/local/bin gravável
  if [ -d "$HOME/.local/bin" ] || mkdir -p "$HOME/.local/bin" 2>/dev/null; then
    # Se /usr/local/bin é gravável e está no PATH, prefere ele? Mantém ~/.local/bin para não pedir sudo
    echo "$HOME/.local/bin"
  elif [ -w "/usr/local/bin" ]; then
    echo "/usr/local/bin"
  else
    echo "$HOME/.local/bin"
  fi
}

# Verifica provider (bore é padrão agora)
check_provider() {
  if command -v bore >/dev/null 2>&1; then
    info "bore encontrado: $(bore --version 2>&1 | head -1)"
  else
    warn "bore não encontrado. Instale o provider padrão:"
    echo "      cargo install bore-cli"
    echo "      # ou brew install bore"
    echo "      # https://github.com/ekzhang/bore"
  fi
  if command -v cloudflared >/dev/null 2>&1; then
    info "cloudflared encontrado (fallback): $(cloudflared --version 2>&1 | head -1)"
  else
    warn "cloudflared não encontrado (opcional, fallback): brew install cloudflared"
  fi
}

main() {
  PLATFORM="$(detect_platform)"
  GOOS="${PLATFORM%%-*}"
  GOARCH="${PLATFORM##*-}"

  if [ "$VERSION" = "latest" ]; then
    URL="https://github.com/${REPO}/releases/latest/download/${BINARY}-${PLATFORM}"
  else
    # permite v0.1.0 ou 0.1.0
    case "$VERSION" in
      v*) TAG="$VERSION" ;;
      *) TAG="v$VERSION" ;;
    esac
    URL="https://github.com/${REPO}/releases/download/${TAG}/${BINARY}-${PLATFORM}"
  fi

  DST_DIR="$(choose_install_dir)"
  mkdir -p "$DST_DIR"
  DST="$DST_DIR/$BINARY"
  TMP="$(mktemp -d)"
  trap 'rm -rf "$TMP"' EXIT INT TERM

  info "Baixando $BINARY $VERSION para $PLATFORM..."
  echo "     $URL"

  if command -v curl >/dev/null 2>&1; then
    if ! curl -fsSL "$URL" -o "$TMP/$BINARY"; then
      err "Falha ao baixar $URL"
      echo "  Verifique se a release $VERSION existe em https://github.com/${REPO}/releases"
      echo "  Alternativa: go install github.com/${REPO}/cmd/envia@latest"
      exit 1
    fi
  elif command -v wget >/dev/null 2>&1; then
    if ! wget -qO "$TMP/$BINARY" "$URL"; then
      err "Falha ao baixar $URL"
      exit 1
    fi
  else
    err "curl ou wget necessário"
    exit 1
  fi

  chmod +x "$TMP/$BINARY"
  # move
  if [ -w "$DST_DIR" ]; then
    mv "$TMP/$BINARY" "$DST"
  else
    warn "Sem permissão em $DST_DIR, tentando com sudo..."
    sudo mv "$TMP/$BINARY" "$DST"
  fi

  info "Instalado em $DST"
  # verifica PATH
  case ":$PATH:" in
    *":$DST_DIR:"*) ;;
    *)
      warn "$DST_DIR não está no PATH"
      echo "  Adicione ao seu shell:"
      echo "    echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc && source ~/.zshrc"
      echo "  ou use caminho completo: $DST"
      ;;
  esac

  # versão
  if "$DST" --version >/dev/null 2>&1; then
    info "$("$DST" --version)"
  fi

  check_provider

  echo ""
  info "Pronto! Teste:"
  echo "    $BINARY --help"
  echo "    $BINARY testes/documento-teste.txt"
  echo "    $BINARY  # modo interativo"
}

main "$@"
