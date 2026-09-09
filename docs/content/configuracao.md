# Configuração

Arquivo salvo em `~/Library/Application Support/envia/config.json` (macOS) ou `~/.config/envia/config.json` (Linux):

```json
{
  "show_hidden": false,
  "provider": "bore"
}
```

- `show_hidden`: mostra `.files` no picker
- `provider`: padrão para `--provider auto`. Vazio = `bore`. Avançado: `"cloudflare"` para usar `cloudflared` mesmo em `auto` (ex: quem tem domínio próprio).

```bash
envia --config              # ver
envia --config ocultos      # toggle (esconde ↔ mostra .files)
envia --config provider=cloudflare  # usa cloudflared em auto
envia --config provider=bore        # volta para bore
envia --config provider=auto        # reseta para padrão (bore)
```

Você também pode editar o JSON diretamente:

```bash
cat ~/Library/Application\ Support/envia/config.json
# ou
cat ~/.config/envia/config.json
```

Futuras flags (`--config` expansível) vão usar o mesmo arquivo.
