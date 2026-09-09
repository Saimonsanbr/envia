package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"envia/internal/app"
	"envia/internal/tunnel"
)

const version = "v0.2.1"

func main() {
	var (
		providerFlag string
		showVersion  bool
		configFlag   string
	)

	flag.StringVar(&providerFlag, "provider", "auto", "provider de túnel: auto (bore → serveo → localhost.run), bore, serveo, localhost.run, cloudflare")
	flag.StringVar(&configFlag, "config", "", "configuração: ocultos (alterna exibição de arquivos ocultos)")
	flag.BoolVar(&showVersion, "version", false, "mostra versão")
	flag.BoolVar(&showVersion, "v", false, "mostra versão (alias)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "envia %s\n\n", version)
		fmt.Fprintf(os.Stderr, "Compartilhe um arquivo diretamente do seu computador.\n\n")
		fmt.Fprintf(os.Stderr, "Uso:\n  envia [opções] [arquivo]\n\n")
		fmt.Fprintf(os.Stderr, "Opções:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExemplos:\n  envia video.mp4\n  envia --provider bore video.mp4\n  envia\n  envia --config ocultos\n")
	}

	// Handle `envia --config` without value (show current config)
	for _, a := range os.Args[1:] {
		if a == "--config" {
			// check if next arg exists and is not a flag; flag.Parse will handle value case
			// if --config is standalone, show config
			hasValue := false
			for i, v := range os.Args {
				if v == "--config" && i+1 < len(os.Args) && !strings.HasPrefix(os.Args[i+1], "-") {
					hasValue = true
					break
				}
				if strings.HasPrefix(v, "--config=") {
					hasValue = true
					break
				}
			}
			if !hasValue {
				if err := app.HandleConfigFlag(""); err != nil {
					fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
					os.Exit(1)
				}
				os.Exit(0)
			}
			break
		}
	}

	flag.Parse()

	if showVersion {
		fmt.Printf("envia %s\n", version)
		os.Exit(0)
	}

	if configFlag != "" {
		// allow --config show / --config ocultos
		if err := app.HandleConfigFlag(configFlag); err != nil {
			fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	args := flag.Args()

	// Validate provider (inclui fallbacks ssh)
	provider := tunnel.Provider(strings.ToLower(providerFlag))
	switch provider {
	case tunnel.ProviderAuto, tunnel.ProviderCloudflare, tunnel.ProviderBore, tunnel.ProviderServeo, tunnel.ProviderLocalhostRun:
	default:
		fmt.Fprintf(os.Stderr, "Erro: provider inválido: %s (use auto, bore, serveo, localhost.run ou cloudflare)\n", providerFlag)
		os.Exit(1)
	}
	// Se auto e config tem provider preferido (avançado), usa ele
	if provider == tunnel.ProviderAuto {
		if cfg, err := app.LoadConfig(); err == nil && cfg.Provider != "" {
			p := tunnel.Provider(strings.ToLower(cfg.Provider))
			switch p {
			case tunnel.ProviderBore, tunnel.ProviderCloudflare, tunnel.ProviderServeo, tunnel.ProviderLocalhostRun:
				provider = p
			}
		}
	}

	var fileArg string
	if len(args) == 0 {
		fileArg = ""
	} else if len(args) == 1 {
		fileArg = args[0]
	} else {
		fmt.Fprintf(os.Stderr, "Erro: informe apenas um arquivo por vez\n")
		flag.Usage()
		os.Exit(1)
	}

	// Context with signal
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.RunUI(ctx, fileArg, provider, webFiles); err != nil {
		if ctx.Err() != nil {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
		os.Exit(1)
	}
}
