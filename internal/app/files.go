package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type fileItem struct {
	name string
	path string
	size string
}

func (f fileItem) FilterValue() string { return f.name }
func (f fileItem) Title() string       { return f.name }
func (f fileItem) Description() string { return f.size }

// ListFiles returns sorted list of files in dir (non-directories).
// It respects config ShowHidden: if false, dotfiles are hidden.
func ListFiles(dir string) ([]string, error) {
	c, _ := LoadConfig()
	return ListFilesWithConfig(dir, c.ShowHidden)
}

// ListFilesWithConfig is same but explicit showHidden flag (for testing).
func ListFilesWithConfig(dir string, showHidden bool) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !showHidden && strings.HasPrefix(name, ".") {
			continue
		}
		files = append(files, name)
	}
	sort.Strings(files)
	return files, nil
}

func humanSize(b int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case b >= GB:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(GB))
	case b >= MB:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(MB))
	case b >= KB:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(KB))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

// PickFile launches interactive selector for current directory.
func PickFile() (string, error) {
	files, err := ListFiles(".")
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", fmt.Errorf("a pasta atual não contém arquivos")
	}

	items := make([]list.Item, 0, len(files))
	for _, f := range files {
		info, err := os.Stat(f)
		var sz string
		if err == nil {
			sz = humanSize(info.Size())
		}
		abs, _ := filepath.Abs(f)
		items = append(items, fileItem{name: f, path: abs, size: sz})
	}

	const defaultWidth = 40
	const defaultHeight = 14

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	l := list.New(items, delegate, defaultWidth, defaultHeight)
	l.Title = "Selecione um arquivo"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)
	l.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	l.Styles.PaginationStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	l.FilterInput.Placeholder = "/ filtrar"

	m := pickerModel{list: l}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}
	pm, ok := finalModel.(pickerModel)
	if !ok {
		return "", fmt.Errorf("erro interno no seletor")
	}
	if pm.cancelled {
		return "", fmt.Errorf("seleção cancelada")
	}
	if pm.choice == "" {
		return "", fmt.Errorf("nenhum arquivo selecionado")
	}
	return pm.choice, nil
}

type pickerModel struct {
	list      list.Model
	choice    string
	cancelled bool
}

func (m pickerModel) Init() tea.Cmd { return nil }

func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit
		case "enter":
			i, ok := m.list.SelectedItem().(fileItem)
			if ok {
				m.choice = i.path
			}
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m pickerModel) View() string {
	// Minimal chrome similar to spec
	header := lipgloss.NewStyle().Bold(true).Render("┌─ envia ──────────────────────┐")
	footer := lipgloss.NewStyle().Faint(true).Render("│ / filtrar  ↑↓ navegar  Enter selecionar  Esc sair │\n└──────────────────────────────────┘")
	return fmt.Sprintf("%s\n%s\n%s", header, m.list.View(), footer)
}

// ResolveFile validates given path and returns absolute path.
func ResolveFile(p string) (string, error) {
	if strings.TrimSpace(p) == "" {
		return "", fmt.Errorf("caminho vazio")
	}
	info, err := os.Stat(p)
	if err != nil {
		return "", fmt.Errorf("arquivo não encontrado: %s", p)
	}
	if info.IsDir() {
		return "", fmt.Errorf("o caminho informado é um diretório.\nEscolha um arquivo")
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	return abs, nil
}
