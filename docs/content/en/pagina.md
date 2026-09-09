# Page and preview

The page is minimal, responsive and works on mobile (no React, no CDN).

- Images (`.jpg`, `.png`, `.webp` etc) with `<img>` at `/preview`
- Videos (`.mp4`, `.webm`) with `<video controls>` and `Range` for seek
- Audios (`.mp3`, `.wav`) with `<audio controls>`
- PDFs with `<iframe src="/preview">`
- Others with file icon + **Download** button (`/download` with `Content-Disposition: attachment`)

Everything comes from the binary — the recipient installs nothing. `preview.css` and `app.js` are served from `embed.FS` and work offline.

## QR code on the page

The page shows `QR code — scan on phone:` with `<img src="/qr" width="180">`.

- `GET /qr` generates PNG on the fly with `qr.GeneratePNG("https://"+r.Host+"/")` (for `bore` uses `http://`).
- Cache `no-cache`, no external dependency, 256x256.

## Endpoints

- `GET /` → HTML with `{{.FileName}}`, `{{.FileSize}}`, `{{.MimeType}}`, `{{.PreviewType}}`
- `GET /preview` → file inline with correct `Content-Type`
- `GET /download` → same file with `Content-Disposition: attachment; filename="..."` (preserves spaces/accents via `filename*`)
- `GET /qr` → PNG of the public link's QR
- `GET /preview.css`, `GET /app.js` → static assets

## Security

- Only serves the chosen file, no `../`
- Headers `nosniff` and `no-referrer`
- Filename with `html/template` escaping and `url.PathEscape` for `filename*`
