# Usage

## Interactive

Lists files in the current folder. Use `↑↓` to navigate, type to filter, `Enter` to select, `Esc` to cancel.

```bash
envia
```

Hidden files (`.DS_Store`, `.env`) are hidden by default. To show:

```bash
envia --config hidden
envia --config  # show current status
```

## Direct

```bash
envia video.mp4
envia ./photo.jpg
envia /Users/saimo/Videos/video.mp4
```

## Choose provider

```bash
envia --provider bore video.mp4        # default, instant
envia --provider serveo video.mp4      # ssh to serveo.net
envia --provider localhost.run video.mp4
envia --provider cloudflare video.mp4  # advanced, needs cloudflared
envia --provider auto video.mp4        # default: bore → serveo → lhr
```

## Config and version

```bash
envia --config              # show current config
envia --config hidden       # toggle .files
envia --config provider=cloudflare  # use cloudflared in auto
envia --config provider=bore

envia --version
envia -v
```

## What appears after choosing

```
╭──────────────────────────────────────────╮
│ File                                     │
│   video.mp4  327.4 MB                     │
│ Server                                   │
│   http://127.0.0.1:43821                  │
│ Tunnel                                   │
│   bore                                   │
╰──────────────────────────────────────────╯
Link
  http://bore.pub:12345

QR code — scan on phone:
 █▀▀▀▀▀█  ▀▀█▄ ...
 █ ███ █ ▄▀██ ...

File stays on your computer.
Waiting for downloads... • Ctrl+C to stop
```

`Ctrl+C` stops server and tunnel — the link dies immediately.

The link outside the box doesn't break when copying in a narrow terminal (tested at 40 cols).
