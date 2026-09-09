# Configuration

File saved at `~/Library/Application Support/envia/config.json` (macOS) or `~/.config/envia/config.json` (Linux):

```json
{
  "show_hidden": false,
  "provider": "bore"
}
```

- `show_hidden`: show `.files` in picker
- `provider`: default for `--provider auto`. Empty = `bore`. Advanced: `"cloudflare"` to use `cloudflared` even in `auto` (e.g., who has own domain).

```bash
envia --config              # show
envia --config hidden       # toggle (hide ↔ show .files)
envia --config provider=cloudflare  # use cloudflared in auto
envia --config provider=bore
envia --config provider=auto        # reset to default (bore)
```

You can also edit the JSON directly:

```bash
cat ~/Library/Application\ Support/envia/config.json
# or
cat ~/.config/envia/config.json
```

Future flags (`--config` expandable) will use the same file.
