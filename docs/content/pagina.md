# Página e preview

A página é minimalista, responsiva e funciona no celular (sem React, sem CDN).

- Imagens (`.jpg`, `.png`, `.webp` etc) com `<img>` em `/preview`
- Vídeos (`.mp4`, `.webm`) com `<video controls>` e `Range` para seek
- Áudios (`.mp3`, `.wav`) com `<audio controls>`
- PDFs com `<iframe src="/preview">`
- Outros com ícone de arquivo + botão **Baixar arquivo** (`/download` com `Content-Disposition: attachment`)

Tudo vem do binário — quem recebe não instala nada. `preview.css` e `app.js` são servidos de `embed.FS` e funcionam offline.

## QR code na página

A página mostra `QR code — escaneie no celular:` com `<img src="/qr" width="180">`.

- `GET /qr` gera PNG na hora com `qr.GeneratePNG("https://"+r.Host+"/")` (para `bore` usa `http://`).
- Cache `no-cache`, sem dependência externa, 256x256.

## Endpoints

- `GET /` → HTML com `{{.FileName}}`, `{{.FileSize}}`, `{{.MimeType}}`, `{{.PreviewType}}`
- `GET /preview` → arquivo inline com `Content-Type` correto
- `GET /download` → mesmo arquivo com `Content-Disposition: attachment; filename="..."` (preserva espaços/acentos via `filename*`)
- `GET /qr` → PNG do QR do link público
- `GET /preview.css`, `GET /app.js` → assets estáticos

## Segurança

- Só serve o arquivo escolhido, sem `../`
- Headers `nosniff` e `no-referrer`
- Nome do arquivo com `html/template` escaping e `url.PathEscape` para `filename*`
