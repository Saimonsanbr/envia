#!/bin/bash
# Grava demo do envia com ffmpeg (sem precisar de vhs/asciinema)
# Uso: ./scripts/record-demo.sh
# Requer: ffmpeg, envia compilado em bin/envia

set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

OUT="demo.mp4"
TMPDIR=$(mktemp -d)
TAPE="$TMPDIR/script.txt"

echo "==> Gravando demo com ffmpeg..."

# Cria um script de terminal que será gravado via 'script' e convertido com ffmpeg
# Usa 'script' para capturar a sessão + ffmpeg para gravar a janela

# Alternativa simples: usa asciinema se disponível, senão usa script + ffmpeg
if command -v vhs >/dev/null 2>&1; then
  echo "vhs encontrado, usando demo.tape..."
  vhs demo.tape
  echo "Gerado demo.gif via vhs"
  exit 0
fi

if command -v asciinema >/dev/null 2>&1 && command -v agg >/dev/null 2>&1; then
  echo "Usando asciinema + agg..."
  asciinema rec --overwrite --command "bash -c 'ls arquivos-engracados/; echo ---; bin/envia arquivos-engracados/windows-pro-mega-blaster-ultra-3000-final.png --provider bore; sleep 2'" "$TMPDIR/demo.cast"
  agg "$TMPDIR/demo.cast" demo.gif
  echo "Gerado demo.gif via asciinema"
  exit 0
fi

# Fallback: usa ffmpeg com x11grab (Linux) ou avfoundation (macOS)
# No macOS, grava a tela com ffmpeg -f avfoundation
echo "Tentando gravar com ffmpeg..."
echo "No macOS, use: brew install vhs && vhs demo.tape"
echo "Ou: brew install asciinema agg && $0"
echo ""
echo "Para gravar manualmente com ffmpeg no macOS:"
echo "  ffmpeg -f avfoundation -i \"1:0\" -r 30 -t 20 demo.mp4  # grava tela"
echo "  # depois converta para gif: ffmpeg -i demo.mp4 -vf \"fps=10,scale=900:-1:flags=lanczos\" demo.gif"
echo ""
echo "Demo tape disponível em: demo.tape"
echo "Conteúdo:"
cat demo.tape
