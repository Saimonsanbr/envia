package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds persistent settings.
type Config struct {
	ShowHidden bool   `json:"show_hidden"`
	// Provider padrão quando --provider auto (vazio = bore). Avançado: "cloudflare" para usar cloudflared mesmo em auto.
	// Permite que usuário com domínio próprio configure cloudflared como padrão editando o JSON.
	Provider string `json:"provider,omitempty"`
}

func configPath() string {
	// Prefer UserConfigDir (~/Library/Application Support on mac, ~/.config on linux)
	if dir, err := os.UserConfigDir(); err == nil && dir != "" {
		return filepath.Join(dir, "envia", "config.json")
	}
	// Fallback to home
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, ".config", "envia", "config.json")
	}
	// Last fallback: current dir
	return filepath.Join(".envia-config.json")
}

// LoadConfig loads config, returns default if not exists.
func LoadConfig() (Config, error) {
	p := configPath()
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{ShowHidden: false}, nil
		}
		return Config{}, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return Config{}, err
	}
	return c, nil
}

// SaveConfig saves config atomically.
func SaveConfig(c Config) error {
	p := configPath()
	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// ToggleHidden flips ShowHidden and persists, returns new value.
func ToggleHidden() (bool, error) {
	c, err := LoadConfig()
	if err != nil {
		return false, err
	}
	c.ShowHidden = !c.ShowHidden
	if err := SaveConfig(c); err != nil {
		return false, err
	}
	return c.ShowHidden, nil
}

// ConfigPath returns path for display/debug.
func ConfigPath() string { return configPath() }

// HandleConfigFlag handles --config ocultos.
func HandleConfigFlag(arg string) error {
	switch arg {
	case "ocultos", "hidden", "oculto", "show_hidden":
		now, err := ToggleHidden()
		if err != nil {
			return err
		}
		if now {
			fmt.Printf("✓ arquivos ocultos agora visíveis\n")
			fmt.Printf("  (config salvo em %s)\n", ConfigPath())
		} else {
			fmt.Printf("✓ arquivos ocultos agora ocultos\n")
			fmt.Printf("  (config salvo em %s)\n", ConfigPath())
		}
		return nil
	case "", "show", "status", "list":
		c, err := LoadConfig()
		if err != nil {
			return err
		}
		fmt.Printf("envia config\n")
		fmt.Printf("  ocultos: %v\n", c.ShowHidden)
		fmt.Printf("  provider: %s\n", func() string {
			if c.Provider == "" {
				return "auto (bore)"
			}
			return c.Provider
		}())
		fmt.Printf("  caminho: %s\n", ConfigPath())
		fmt.Printf("\nUso:\n")
		fmt.Printf("  envia --config ocultos              # alterna .files\n")
		fmt.Printf("  envia --config provider=cloudflare  # usa cloudflared em auto (avançado)\n")
		fmt.Printf("  envia --config provider=bore        # volta para bore\n")
		fmt.Printf("\nEdite o JSON diretamente para provider custom:\n")
		fmt.Printf("  %s\n", ConfigPath())
		return nil
	default:
		// permite --config provider=cloudflare
		if arg == "provider=cloudflare" || arg == "provider=auto" || arg == "provider=bore" || arg == "provider=serveo" || arg == "provider=localhost.run" {
			parts := arg[len("provider="):]
			c, err := LoadConfig()
			if err != nil {
				return err
			}
			if parts == "auto" {
				c.Provider = ""
			} else {
				c.Provider = parts
			}
			if err := SaveConfig(c); err != nil {
				return err
			}
			fmt.Printf("✓ provider padrão agora: %s\n", func() string {
				if c.Provider == "" {
					return "auto (bore)"
				}
				return c.Provider
			}())
			fmt.Printf("  (config salvo em %s)\n", ConfigPath())
			return nil
		}
		return fmt.Errorf("opção de config desconhecida: %s\nUse: --config ocultos, --config provider=cloudflare, --config show", arg)
	}
}
