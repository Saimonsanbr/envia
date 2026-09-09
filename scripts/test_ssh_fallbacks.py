#!/usr/bin/env python3
"""
Testa exaustivamente os fallbacks SSH (serveo e localhost.run) com arquivo vazio.
Gera um arquivo vazio via dd ou usa testes/documento-teste.txt e tenta cada provider
várias vezes, capturando tipo de falha e tempo.

Uso:
  python3 scripts/test_ssh_fallbacks.py
  python3 scripts/test_ssh_fallbacks.py --provider serveo --runs 5
  python3 scripts/test_ssh_fallbacks.py --provider bore --runs 3
"""
import argparse, subprocess, time, re, os, sys, tempfile, pathlib, signal, select, pty

PROVIDERS = ["bore", "serveo", "localhost.run"]
REGEX = {
    "bore": re.compile(r"bore\.pub:\d+"),
    "serveo": re.compile(r"https://[a-z0-9-]+\.serveo\.net"),
    "localhost.run": re.compile(r"https://[a-z0-9-]+\.lhr\.life"),
    "cloudflare": re.compile(r"https://[a-z0-9-]+\.trycloudflare\.com"),
}

def run_once(provider, file_path, timeout=20):
    """Roda envia com provider e captura saída via pty, retorna dict com resultado."""
    master, slave = pty.openpty()
    cmd = ["bin/envia", "--provider", provider, file_path]
    start = time.time()
    proc = subprocess.Popen(cmd, stdin=slave, stdout=slave, stderr=slave, close_fds=True, cwd=".")
    os.close(slave)
    buf = b""
    url = None
    error = None
    ready = False
    try:
        while time.time() - start < timeout:
            r, _, _ = select.select([master], [], [], 0.5)
            if master in r:
                try:
                    data = os.read(master, 4096)
                    if not data:
                        break
                    buf += data
                    # strip ansi for regex
                    clean = re.sub(br"\x1b\[[0-9;]*m", b"", buf)
                    if provider in REGEX and REGEX[provider].search(clean.decode(errors="ignore")):
                        url = REGEX[provider].search(clean.decode(errors="ignore")).group(0)
                    if b"Aguardando" in buf:
                        ready = True
                        break
                    if b"Erro" in buf:
                        # captura erro
                        error = clean.decode(errors="ignore")[-2000:]
                        break
                except:
                    break
            if proc.poll() is not None:
                # processo terminou antes de ready
                break
    finally:
        if proc.poll() is None:
            try:
                proc.send_signal(signal.SIGINT)
                proc.wait(timeout=3)
            except:
                proc.kill()
        # drain
        for _ in range(3):
            r, _, _ = select.select([master], [], [], 0.2)
            if master in r:
                try:
                    data = os.read(master, 4096)
                    buf += data
                except:
                    break
        os.close(master)
    elapsed = time.time() - start
    clean = re.sub(br"\x1b\[[0-9;]*m", b"", buf).decode(errors="ignore")
    # tenta extrair erro
    if not ready and proc.returncode and proc.returncode != 0:
        # pega últimas linhas
        error = clean[-1000:]
    return {
        "provider": provider,
        "url": url,
        "ready": ready,
        "elapsed": round(elapsed, 2),
        "exit_code": proc.returncode,
        "error": error,
        "raw": clean[-500:] if len(clean) > 500 else clean,
    }

def main():
    parser = argparse.ArgumentParser(description="Testa fallbacks SSH exaustivamente")
    parser.add_argument("--provider", choices=PROVIDERS+["auto", "cloudflare"], default="serveo", help="provider a testar")
    parser.add_argument("--runs", type=int, default=5, help="número de tentativas")
    parser.add_argument("--file", default="testes/documento-teste.txt", help="arquivo de teste (vazio ou existente)")
    parser.add_argument("--timeout", type=int, default=20, help="timeout por tentativa (s)")
    args = parser.parse_args()

    # garante arquivo de teste existe (cria vazio se não existir)
    p = pathlib.Path(args.file)
    if not p.exists():
        print(f"Criando arquivo vazio {p}")
        p.parent.mkdir(parents=True, exist_ok=True)
        p.write_bytes(b"")

    print(f"Testando provider={args.provider} runs={args.runs} file={args.file} timeout={args.timeout}s")
    print("="*70)
    results = []
    for i in range(args.runs):
        print(f"\n--- Tentativa {i+1}/{args.runs} ---")
        r = run_once(args.provider, args.file, timeout=args.timeout)
        results.append(r)
        status = "✓ OK" if r["ready"] else "✗ FALHA"
        print(f"{status} url={r['url']} elapsed={r['elapsed']}s exit={r['exit_code']}")
        if r["error"]:
            # mostra apenas as 3 primeiras linhas do erro
            err_lines = r["error"].strip().split("\n")[:5]
            print("  erro:", " | ".join(err_lines))
        # detecta tipo de falha
        if not r["ready"]:
            if "ssh" in (r["error"] or "").lower() or "not found" in (r["error"] or "").lower():
                print("  → falha: ssh não encontrado ou bore não instalado")
            elif "timeout" in (r["error"] or "").lower():
                print("  → falha: timeout ao aguardar URL (15s)")
            elif "429" in (r["error"] or ""):
                print("  → falha: rate limit 429")
            elif "no such host" in (r["error"] or "").lower():
                print("  → falha: DNS")
        time.sleep(1)  # intervalo entre tentativas para não hitar rate limit

    print("\n" + "="*70)
    print("Resumo:")
    ok = sum(1 for r in results if r["ready"])
    print(f"  {ok}/{len(results)} sucessos")
    print(f"  tempos: {[r['elapsed'] for r in results]}")
    # estatísticas
    if ok > 0:
        avg = sum(r["elapsed"] for r in results if r["ready"]) / ok
        print(f"  média OK: {avg:.2f}s")
    # recomendações
    print("\nRecomendações:")
    print("  - bore: deve ser <4s feliz, sem health. Se falhar, verifique 'bore --version' e 'bore.pub' acessível.")
    print("  - serveo: precisa ssh e internet, URL https://*.serveo.net, timeout 15s. Falha comum: ssh não encontrado, firewall bloqueando porta 22.")
    print("  - localhost.run: ssh -R 80:localhost:PORT nokey@localhost.run, URL https://*.lhr.life, precisa ssh.")
    print("  - Para implementar no envia, use regex serveoRegex e localhostRunRegex e capture URL como feito para bore.")

if __name__ == "__main__":
    main()
