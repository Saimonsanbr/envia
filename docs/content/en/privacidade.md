# Privacy

> `envia` **does not store** your file. No upload to its own backend.

But traffic **goes through the tunnel provider** (`bore`/`serveo`/`lhr` or `cloudflared`) — it can see the traffic. The link is temporary and has no auth in v0.2.3: whoever has the URL can access the file while you keep `envia` running.

```
link = access to file
```

There is no `api.envia.com`, database or account. The path is:

```
A → Tunnel (bore.pub / serveo.net / trycloudflare.com) → B
```

We collect nothing: no analytics, no trackers, no account, no history. The tool is stateless.

> Note: although `envia` has no storage, traffic through third-party infra may be visible to the chosen provider. Don't make absolute “total privacy” claims.
