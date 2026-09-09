# Introdução

`envia` transforma temporariamente seu computador em um servidor de arquivos.

Um comando, um link. Quem recebe só abre no navegador — sem cadastro, sem app, sem upload para nuvem do `envia`.

```bash
envia video.mp4
# → http://bore.pub:12345
```

O arquivo continua no seu disco. Seu PC abre um servidor em `127.0.0.1` com porta aleatória, um túnel público (`bore`) expõe esse servidor, você compartilha a URL e a pessoa baixa direto de você. Quando você dá `Ctrl+C`, o link morre.

```
arquivo → servidor local → túnel → navegador
 (disco)    (127.0.0.1)   (bore.pub)  (preview/download)
```

> Sem backend próprio, sem conta, sem painel. Só um binário.

## Aviso

Projeto entusiasta, feito no tempo livre. Funciona, mas ainda pode ter bugs.

- **Testado em macOS (M1) e Linux** — ambos funcionando perfeitamente com `bore`.
- **Windows em breve** — será testado logo logo, com `bore.exe` bundle.
- A ideia é manter tudo simples e rápido. A v0.2.3 ainda é MVP de um único arquivo por vez.

Se encontrar algo estranho, abra uma issue em https://github.com/Saimonsanbr/envia/issues
