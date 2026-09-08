package app

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"

	"envia/internal/httpserver"
	"envia/internal/tunnel"
)

const (
	statePicking = iota
	stateLoading
	stateReady
	stateError
)

type tunnelErrMsg struct{ err error }

type appModel struct {
	state     int
	list      list.Model
	spinner   spinner.Model
	filePath  string
	fileName  string
	fileSize  string
	addr      string
	publicURL string
	provider  tunnel.Provider
	webFS     fs.FS
	ctx       context.Context
	cancel    context.CancelFunc
	err       error
	mgr       *tunnel.Manager
	srv       *httpserver.Server
	width     int
	height    int
	// for picking transition, we keep initial arg presence
	hasInitialFile bool
}

func newAppModel(ctx context.Context, fileArg string, provider tunnel.Provider, webFS fs.FS) (*appModel, error) {
	spin := spinner.New()
	spin.Spinner = spinner.Dot
	spin.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

	m := &appModel{
		spinner:  spin,
		provider: provider,
		webFS:    webFS,
		ctx:      ctx,
	}

	if fileArg != "" {
		abs, err := ResolveFile(fileArg)
		if err != nil {
			return nil, err
		}
		info, _ := os.Stat(abs)
		var sz string
		if info != nil {
			sz = humanSize(info.Size())
		}
		m.filePath = abs
		m.fileName = filepath.Base(abs)
		m.fileSize = sz
		m.hasInitialFile = true
		m.state = stateLoading
	} else {
		// need picker
		files, err := ListFiles(".")
		if err != nil {
			return nil, err
		}
		if len(files) == 0 {
			return nil, fmt.Errorf("a pasta atual não contém arquivos")
		}
		items := make([]list.Item, 0, len(files))
		for _, f := range files {
			info, _ := os.Stat(f)
			var sz string
			if info != nil {
				sz = humanSize(info.Size())
			}
			abs, _ := filepath.Abs(f)
			items = append(items, fileItem{name: f, path: abs, size: sz})
		}
		delegate := list.NewDefaultDelegate()
		delegate.ShowDescription = true
		l := list.New(items, delegate, 50, 14)
		l.Title = "Selecione um arquivo"
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(true)
		l.SetShowHelp(true)
		l.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
		l.Styles.PaginationStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		l.FilterInput.Placeholder = "/ filtrar"
		m.list = l
		m.state = statePicking
	}

	return m, nil
}

func (m appModel) Init() tea.Cmd {
	if m.state == statePicking {
		return nil
	}
	if m.state == stateLoading {
		return tea.Batch(m.spinner.Tick, m.startTunnelCmd())
	}
	return nil
}

type tunnelReadyMsg struct {
	url      string
	provider tunnel.Provider
	addr     string
	mgr      *tunnel.Manager
	srv      *httpserver.Server
}

func (m *appModel) startTunnelCmd() tea.Cmd {
	filePath := m.filePath
	provider := m.provider
	webFS := m.webFS
	ctx := m.ctx
	return func() tea.Msg {
		srv, err := httpserver.New(filePath, webFS)
		if err != nil {
			return tunnelErrMsg{err}
		}
		addr, err := srv.Listen()
		if err != nil {
			return tunnelErrMsg{err}
		}
		mgr := &tunnel.Manager{}
		cfg := tunnel.Config{Addr: addr, Provider: provider}
		url, err := mgr.StartWithRetry(ctx, cfg)
		if err != nil {
			_ = srv.Close()
			return tunnelErrMsg{err}
		}
		return tunnelReadyMsg{url: url, provider: mgr.Provider(), addr: addr, mgr: mgr, srv: srv}
	}
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.state == statePicking && msg.Width > 0 && msg.Height > 4 {
			m.list.SetWidth(msg.Width)
			m.list.SetHeight(msg.Height - 4)
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.srv != nil {
				_ = m.srv.Close()
			}
			if m.mgr != nil {
				_ = m.mgr.Close()
			}
			return m, tea.Quit
		case "esc":
			if m.state == statePicking {
				m.err = fmt.Errorf("seleção cancelada")
				return m, tea.Quit
			}
			// in loading/ready, esc quits as well
			if m.srv != nil {
				_ = m.srv.Close()
			}
			if m.mgr != nil {
				_ = m.mgr.Close()
			}
			return m, tea.Quit
		case "enter":
			if m.state == statePicking {
				i, ok := m.list.SelectedItem().(fileItem)
				if ok {
					m.filePath = i.path
					m.fileName = i.name
					m.fileSize = i.size
					m.state = stateLoading
					return m, tea.Batch(m.spinner.Tick, m.startTunnelCmd())
				}
			}
		}

	case spinner.TickMsg:
		if m.state == stateLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case tunnelReadyMsg:
		m.publicURL = msg.url
		m.addr = msg.addr
		m.mgr = msg.mgr
		m.srv = msg.srv
		if msg.provider != "" {
			m.provider = msg.provider
		}
		m.state = stateReady
		return m, nil

	case tunnelErrMsg:
		m.err = msg.err
		m.state = stateError
		return m, tea.Quit
	}

	// delegate to list when picking
	if m.state == statePicking {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m appModel) View() string {
	switch m.state {
	case statePicking:
		header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).Render("envia v0.1.0")
		sub := lipgloss.NewStyle().Faint(true).Render("Compartilhe um arquivo diretamente do seu computador.")
		by := lipgloss.NewStyle().Faint(true).Render("by @indigena.dev")
		top := lipgloss.JoinVertical(lipgloss.Center, header, sub, by)
		footer := lipgloss.NewStyle().Faint(true).Render("│ / filtrar  ↑↓ navegar  Enter selecionar  Esc sair │")
		return lipgloss.JoinVertical(lipgloss.Left, top, "", m.list.View(), footer)

	case stateLoading:
		// Header same as picking for consistency
		header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).Render("envia v0.1.0")
		sub := lipgloss.NewStyle().Faint(true).Render("Compartilhe um arquivo diretamente do seu computador.")
		by := lipgloss.NewStyle().Faint(true).Render("by @indigena.dev")
		top := lipgloss.JoinVertical(lipgloss.Center, header, sub, by)

		// File box
		boxStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("212")).Padding(1, 2).Width(50)
		fileBox := boxStyle.Render(fmt.Sprintf("%s\n%s", lipgloss.NewStyle().Bold(true).Render(m.fileName), lipgloss.NewStyle().Faint(true).Render(m.fileSize)))

		sp := lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Render(m.spinner.View())
		loading := lipgloss.JoinHorizontal(lipgloss.Center, sp, "  Criando link...")

		// keep subtle, no attempt details
		provName := string(m.provider)
		if m.provider == tunnel.ProviderAuto {
			provName = "auto"
		}
		hint := lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("túnel: %s  •  %s", provName, m.addr))

		return lipgloss.JoinVertical(lipgloss.Left, top, "", fileBox, "", loading, hint)

	case stateReady:
		header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).Render("envia v0.1.0")
		sub := lipgloss.NewStyle().Faint(true).Render("Compartilhe um arquivo diretamente do seu computador.")
		by := lipgloss.NewStyle().Faint(true).Render("by @indigena.dev")
		top := lipgloss.JoinVertical(lipgloss.Center, header, sub, by)

		boxStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("212")).Padding(1, 2).Width(60)
		provLabel := string(m.provider)
		if m.provider == "" {
			provLabel = "cloudflare"
		}
		content := fmt.Sprintf(
			"%s\n%s\n\n%s\n  %s\n\n%s\n  %s\n\n%s\n  %s",
			lipgloss.NewStyle().Bold(true).Render("Arquivo"),
			fmt.Sprintf("  %s  %s", m.fileName, lipgloss.NewStyle().Faint(true).Render(m.fileSize)),
			lipgloss.NewStyle().Bold(true).Render("Servidor"),
			"http://"+m.addr,
			lipgloss.NewStyle().Bold(true).Render("Túnel"),
			provLabel,
			lipgloss.NewStyle().Bold(true).Render("Link público"),
			lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true).Render(m.publicURL),
		)
		box := boxStyle.Render(content)

		foot := lipgloss.NewStyle().Faint(true).Render("O arquivo continua no seu computador.\nNenhum upload foi feito para o envia.")
		wait := lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Render("Aguardando downloads...  •  Ctrl+C para encerrar")
		return lipgloss.JoinVertical(lipgloss.Left, top, "", box, "", foot, "", wait)

	case stateError:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(fmt.Sprintf("Erro: %v", m.err))
	}
	return ""
}

func isTerminal() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) && isatty.IsTerminal(os.Stdin.Fd())
}

// RunUI is the unified entry point with bubbletea UI.
func RunUI(ctx context.Context, fileArg string, provider tunnel.Provider, webFS fs.FS) error {
	// Fallback to plain mode if not a TTY (pipes, CI, tests)
	if !isTerminal() {
		var filePath string
		if fileArg != "" {
			filePath = fileArg
		} else {
			// picker needs TTY
			return fmt.Errorf("selecione um arquivo: use envia <arquivo> quando não estiver em um terminal interativo")
		}
		return Run(ctx, filePath, provider, webFS)
	}

	model, err := newAppModel(ctx, fileArg, provider, webFS)
	if err != nil {
		return err
	}

	// For direct file mode, we still want to use the bubbletea program (spinner) rather than plain prints.
	// The model already is in loading state.
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithContext(ctx))

	finalModel, err := p.Run()
	if err != nil {
		return err
	}
	am, ok := finalModel.(appModel)
	if !ok {
		return fmt.Errorf("erro interno de UI")
	}
	if am.state == stateError {
		return am.err
	}
	if am.state == stateReady {
		// After tea quits (user Ctrl+C), we already closed in Update, but ensure cleanup and wait for ctx?
		// Block until ctx done was previously in app.Run; now tea handles quit via key, but also signal should quit.
		// If we reached ready, we need to keep running until signal or quit.
		// However tea already quit on ctrl+c; we just need to ensure tunnel closed.
		// If quit was via signal, ctx will be cancelled and we return.
		// For ready state, we actually want to stay alive showing ready screen until Ctrl+C.
		// Our current flow quits immediately after reaching ready, which is wrong.
		// Instead we should not quit on ready; we should stay in tea until user quits or ctx cancelled.
		// The above Run returns immediately after ready, so we need to adjust: don't return on tunnelReadyMsg, stay.
		// But we already returned ready state; the caller will see ready and then need to block?
		// To keep alive, we should not have quit; the tea program should remain running in ready state until user quits.
		// So the Run should have blocked until user pressed Ctrl+C inside tea.
		// That means finalModel will be after user quit, not immediately after tunnelReady.
		// Therefore if we get here with stateReady, it means user already quit gracefully.
		if am.mgr != nil {
			_ = am.mgr.Close()
		}
		if am.srv != nil {
			_ = am.srv.Close()
		}
		return nil
	}
	if am.err != nil {
		return am.err
	}
	// If we were in picking and cancelled, error already handled
	return nil
}

// Run is legacy plain mode (kept for tests); now delegates to RunUI if needed.
