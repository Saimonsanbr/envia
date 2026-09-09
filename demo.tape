Output demo.gif
Set FontSize 18
Set Width 900
Set Height 700
Set Padding 20
Set Theme "Catppuccin Mocha"
Set Shell "zsh"
Set FontFamily "JetBrains Mono"

Hide
Type "clear" Enter
Sleep 500ms
Show

# Intro - mostra os arquivos engraçados
Type "ls arquivos-engracados/ | head -20" Enter
Sleep 1s
Type "echo '— 13 arquivos prontos para o envia —'" Enter
Sleep 800ms

# Demo 1: arquivo de imagem (preview)
Type "bin/envia arquivos-engracados/windows-pro-mega-blaster-ultra-3000-final.png --provider bore" Enter
Sleep 6s

# Encerra
Ctrl+C
Sleep 1s

# Demo 2: modo interativo (picker)
Type "clear" Enter
Sleep 500ms
Type "bin/envia" Enter
Sleep 1.5s
Type "iphone54" Sleep 300ms
Sleep 800ms
Enter
Sleep 6s
Ctrl+C
Sleep 1s

Type "echo 'envia v0.2.3 — um arquivo, um link, um QR. sem nuvem.'" Enter
Sleep 1s
