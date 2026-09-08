package app

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"envia/internal/httpserver"
	"envia/internal/tunnel"
)

// Run coordinates server + tunnel and blocks until context cancelled.
func Run(ctx context.Context, filePath string, provider tunnel.Provider, webFS fs.FS) error {
	abs, err := ResolveFile(filePath)
	if err != nil {
		return err
	}

	info, err := os.Stat(abs)
	if err != nil {
		return fmt.Errorf("falha ao ler arquivo: %w", err)
	}

	// Create server
	srv, err := httpserver.New(abs, webFS)
	if err != nil {
		return err
	}

	addr, err := srv.Listen()
	if err != nil {
		return err
	}
	defer srv.Close()

	// Start tunnel
	mgr := &tunnel.Manager{}
	cfg := tunnel.Config{
		Addr:     addr,
		Provider: provider,
	}

	fmt.Printf("\nArquivo\n  %s\n  %s\n\n", filepath.Base(abs), humanSize(info.Size()))
	fmt.Printf("Servidor\n  http://%s\n\n", addr)

	// Show tunnel provider name
	provName := string(provider)
	if provider == tunnel.ProviderAuto {
		provName = "auto (cloudflare → bore)"
	}
	fmt.Printf("Túnel\n  %s\n\n", provName)
	fmt.Printf("Criando link...\n")

	publicURL, err := mgr.Start(ctx, cfg)
	if err != nil {
		return err
	}
	defer mgr.Close()

	fmt.Printf("\nLink público\n  %s\n\n", publicURL)
	fmt.Printf("O arquivo continua no seu computador.\nNenhum upload foi feito para o envia.\n\n")
	fmt.Printf("Aguardando downloads...\n\n")
	fmt.Printf("Pressione Ctrl+C para encerrar.\n")

	// Wait for context cancellation
	<-ctx.Done()

	fmt.Printf("\nEncerrando...\n")
	_ = mgr.Close()
	_ = srv.Close()
	return nil
}
