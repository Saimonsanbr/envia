# Uso

## Interativo

Lista arquivos da pasta atual. Use `↑↓` para navegar, digite para filtrar, `Enter` para selecionar, `Esc` para cancelar.

```bash
envia
```

Arquivos ocultos (`.DS_Store`, `.env`) ficam escondidos por padrão. Para mostrar:

```bash
envia --config ocultos
envia --config  # ver status
```

## Direto

```bash
envia video.mp4
envia ./foto.jpg
envia /Users/saimo/Videos/video.mp4
```

## Escolher provider

```bash
envia --provider bore video.mp4        # padrão, instantâneo
envia --provider serveo video.mp4      # ssh para serveo.net
envia --provider localhost.run video.mp4
envia --provider cloudflare video.mp4  # avançado, precisa cloudflared
envia --provider auto video.mp4        # padrão: bore → serveo → lhr
```

## Config e versão

```bash
envia --config              # mostra config atual
envia --config ocultos      # alterna .files
envia --config provider=cloudflare  # usa cloudflared em auto (avançado)
envia --config provider=bore

envia --version
envia -v
```

## O que aparece depois de escolher

```
╭──────────────────────────────────────────╮
│ Arquivo                                  │
│   video.mp4  327.4 MB                     │
│ Servidor                                 │
│   http://127.0.0.1:43821                  │
│ Túnel                                    │
│   bore                                   │
╰──────────────────────────────────────────╯
Link público
  http://bore.pub:12345

QR code — escaneie no celular:
 █▀▀▀▀▀█  ▀▀█▄ ...
 █ ███ █ ▄▀██ ...

O arquivo continua no seu computador.
Aguardando downloads... • Ctrl+C para encerrar
```

`Ctrl+C` encerra servidor e túnel — o link para de funcionar na hora.

O link fora da box não quebra ao copiar em terminal estreito (testado em 40 cols).
