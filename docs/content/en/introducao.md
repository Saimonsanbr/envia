# Introduction

`envia` temporarily turns your computer into a file server.

One command, one link. The recipient just opens it in the browser — no signup, no app, no upload to `envia`'s cloud.

```bash
envia video.mp4
# → http://bore.pub:12345
```

The file stays on your disk. Your PC opens a server on `127.0.0.1` with a random port, a public tunnel (`bore`) exposes that server, you share the URL and the person downloads directly from you. When you press `Ctrl+C`, the link dies.

```
file → local server → tunnel → browser
 (disk)   (127.0.0.1)   (bore.pub)  (preview/download)
```

> No backend, no account, no dashboard. Just a binary.

## Notice

Hobby project, built in spare time. It works, but may still have bugs.

- **Tested on macOS (M1) and Linux** — both work perfectly with `bore`.
- **Windows coming soon** — will be tested shortly with `bore.exe` bundle.
- The idea is to keep everything simple and fast. v0.2.3 is still an MVP for a single file at a time.

If you find something weird, open an issue at https://github.com/Saimonsanbr/envia/issues
