# Privacidade

> O `envia` **não armazena** seu arquivo. Não tem upload para backend próprio.

Mas o tráfego **passa pelo provider de túnel** (`bore`/`serveo`/`lhr` ou `cloudflared`) — ele pode ver o tráfego. O link é temporário e sem autenticação na v0.2.3: quem tiver a URL acessa o arquivo enquanto você deixar o `envia` rodando.

```
link = acesso ao arquivo
```

Não existe `api.envia.com`, banco ou conta. O caminho é:

```
A → Tunnel (bore.pub / serveo.net / trycloudflare.com) → B
```

Não coletamos nada: sem analytics, sem trackers, sem conta, sem histórico. A ferramenta é stateless.

> Nota: embora o `envia` não tenha storage, o tráfego pela infra de terceiros pode ser visível para o provider escolhido. Não faça afirmações de “privacidade total”.
