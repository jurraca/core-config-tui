package main

import (
	"fmt"
  "log"
	"os"
	"strconv"
	"strings"
  "text/template"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

const maxWidth = 100

type Config struct {
	DataDir           string
	Network           string
	RPCAuth           string
	RPCBind           string
	RPCPort           string
	RPCAllowIP        string
	Server            int 
	MaxConnections    string
	TxIndex           int
	Prune             string
	IncludeConf       string
	LoadBlock         string
	MaxMempool        string
	MaxOrphanTx       string
	MempoolExpiry     string
	Par               string
	PersistMempool    int
	PersistMempoolV1  int
	Pid               string
	Reindex           int
	ReindexChainstate int
	Settings          string
	ShutdownNotify    string
	StartupNotify     string
}

var (
	datadir           = "~/.bitcoin"
	network           = "main"
	server            = false 
	rpcauth           string
	rpcport           string
	rpcallowip        string
	rpcbind           string
	maxconnections    string
	includeConf       string
	loadBlock         string
	maxMempool        string
	maxOrphanTx       string
	mempoolExpiry     string
	par               string
	persistMempool    bool
	persistMempoolV1  bool
	pid               string
	prune             string
	reindex           bool
	reindexChainstate bool
	settings          string
	shutdownNotify    string
	startupNotify     string
	txindex           = false

	red    = lipgloss.AdaptiveColor{Light: "#FE5F86", Dark: "#FE5F86"}
	indigo = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
	green  = lipgloss.AdaptiveColor{Light: "#02BA84", Dark: "#02BF87"}
	orange = lipgloss.AdaptiveColor{Light: "#FFA500", Dark: "#FF8C00"}
)```

type Styles struct {
	Base,
	HeaderText,
	Status,
	StatusHeader,
	Highlight,
	ErrorHeaderText,
	Help lipgloss.Style
}

func NewStyles(lg *lipgloss.Renderer) *Styles {
	s := Styles{}
	s.Base = lg.NewStyle().
		Padding(1, 4, 0, 1)
	s.HeaderText = lg.NewStyle().
		Foreground(indigo).
		Bold(true).
		Padding(0, 1, 0, 2)
	s.Status = lg.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(indigo).
		PaddingLeft(1).
		MarginTop(1)
	s.StatusHeader = lg.NewStyle().
		Foreground(green).
		Bold(true)
	s.Highlight = lg.NewStyle().
		Foreground(lipgloss.Color("208"))
	s.ErrorHeaderText = s.HeaderText.
		Foreground(red)
	s.Help = lg.NewStyle().
		Foreground(lipgloss.Color("240"))
	return &s
}

type state int

const (
	statusNormal state = iota
	stateDone
)

type Model struct {
	state  state
	lg     *lipgloss.Renderer
	styles *Styles
	form   *huh.Form
	width  int
}

func NewModel() Model {
	m := Model{width: maxWidth}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("datadir").
				Title("Data Directory").
				Description("Directory to store blockchain data\n(defaults to ~/.bitcoin)").
				Value(&datadir),

			huh.NewSelect[string]().
				Key("chain").
				Title("Select Network").
				Description("The network to run bitcoin on.").
				Options(
					huh.NewOption("mainnet", "main"),
					huh.NewOption("testnet", "test"),
					huh.NewOption("testnet4", "testnet4"),
					huh.NewOption("regtest", "regtest"),
					huh.NewOption("signet", "signet"),
				).
				Value(&network),

			huh.NewConfirm().
				Key("txindex").
				Title("Transaction Index").
				Description("Maintain a full transaction index,\nused by the getrawtransaction rpc\n call (default: No)").
				Value(&txindex),
      
			huh.NewInput().
				Key("prune").
				Title("Prune").
				Description("Prune the blockchain database.\n Possible values: \n 0 = disable pruning blocks (default),\n 1 = allow manual pruning via RPC,\n >=550 = automatically prune block files\n to stay under the specified target size in MiB").
				Validate(func(v string) error {
					if txindex == true && v != "0" {
						return fmt.Errorf("pruning is incompatible with txindex. If you want to use pruning, you must disable txindex.")
					}
					return nil
				}).
				Value(&prune),
		).Title("Basics"),
    huh.NewGroup(
      huh.NewConfirm().
        Key("server").
        Title("Enable RPC Server").
        Description("Accept command line and JSON-RPC commands. (default: false)").
        Value(&server),
			),
		huh.NewGroup(
      huh.NewInput().
        Key("rpcauth").
        Title("RPC Auth").
        Description("Auth for JSON-RPC connections.\n Username and HMAC-SHA-256 hashed password for JSON-RPC connections.\n See the canonical python script included in share/rpcauth to generate this value.").
        Value(&rpcauth).
        Hidden(func() bool { return !server }),
      huh.NewInput().
        Key("rpcport").
        Title("RPC Port").
        Description("Port for RPC connections (default: 8332)").
        Value(&rpcport).
        Hidden(func() bool { return !server }),
      huh.NewInput().
        Key("rpcallowip").
        Title("RPC Allow IP").
        Description("Allow JSON-RPC connections from specified source.").
        Value(&rpcallowip).
        Hidden(func() bool { return !server }),
      huh.NewInput().
        Key("rpcbind").
        Title("RPC bind").
        Description("Bind to given address to listen for JSON-RPC connections.").
        Value(&rpcbind).
        Hidden(func() bool { return !server }),
    ).withHideFunc().
		Title("RPCs"),
	).WithWidth(55).
		WithShowHelp(false).
		WithShowErrors(false)
	return m
}

func (m Model) Init() tea.Cmd {
	return m.form.Init()
}

func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = min(msg.Width, maxWidth) - m.styles.Base.GetHorizontalFrameSize()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc", "q":
			return m, tea.Quit
		}
	}

	var cmds []tea.Cmd

	// Process the form
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	if m.form.State == huh.StateCompleted {
		// Quit when the form is done.
		cmds = append(cmds, tea.Quit)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	s := m.styles

	switch m.form.State {
	case huh.StateCompleted:
    return m.CompletedMsg(*s)

	default:
    // Status (right side)
    var status string
    status = m.StatusBar(*s, m.form, status)
		// Form (left side)
		v := strings.TrimSuffix(m.form.View(), "\n\n")
		form := m.lg.NewStyle().Margin(1, 0).Render(v)

		errors := m.form.Errors()
		header := m.appBoundaryView("Bitcoin Core Configuration")
		if len(errors) > 0 {
			header = m.appErrorBoundaryView(m.errorView())
		}
		body := lipgloss.JoinHorizontal(lipgloss.Top, form, status, "\n")

		footer := m.appBoundaryView(m.form.Help().ShortHelpView(m.form.KeyBinds()))
		if len(errors) > 0 {
			footer = m.appErrorBoundaryView("")
		}

		return s.Base.Render(header + "\n" + body + "\n\n" + footer)
	}
}

func (m Model) errorView() string {
	var s string
	for _, err := range m.form.Errors() {
		s += err.Error()
	}
	return s
}

func (m Model) appBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.HeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(indigo),
	)
}

func (m Model) appErrorBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.ErrorHeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(red),
	)
}

func (m Model) StatusBar(s Styles, form *huh.Form, status string) string {
    var (
      chain   string
      txindex string
      server string
      rpcauth string
      rpcport string
      rpcallowip string
      rpcbind string
    )

    if m.form.GetString("chain") != "" {
      chain = "Network: " + m.form.GetString("chain") + "\n"
    }
    if m.form.GetBool("txindex") != false {
      txindex = "TxIndex enabled: " + strconv.FormatBool(m.form.GetBool("txindex")) + "\n"
    }
    if m.form.GetBool("server") != false {
      server = "RPC server enabled: " + strconv.FormatBool(m.form.GetBool("server")) + "\n"
    }
    if m.form.GetString("rpcauth") != "" {
      rpcauth = "RPC auth: " + m.form.GetString("rpcauth") + "\n"
    }
    if m.form.GetString("rpcport") != "" {
      rpcport = "RPC port: " + m.form.GetString("rpcport") + "\n"
    }
    if m.form.GetString("rpcallowip") != "" {
      rpcallowip = "RPC allow IP: " + m.form.GetString("rpcallowip") + "\n"
    }
    if m.form.GetString("rpcbind") != "" {
      rpcbind = "RPC bind: " + m.form.GetString("rpcbind") + "\n"
    }
    const statusWidth = 28
    statusMarginLeft := m.width - statusWidth - lipgloss.Width(form.View()) - s.Status.GetMarginRight()
    return s.Status.
      Height(lipgloss.Height(form.View())).
      Width(statusWidth).
      MarginLeft(statusMarginLeft).
      Render(s.StatusHeader.Render("Current Config") + 
        "\n\n" +
        "Datadir: " + m.form.GetString("datadir") + "\n" +
        chain +
        txindex +
        server + 
        rpcauth +
        rpcport +
        rpcallowip +
        rpcbind + "\n",
      )
  }
  
func (m Model) CompletedMsg(s Styles) string {
  var (
    b          strings.Builder
    chainVal   string
    datadirVal string
  )
  chainVal = s.Highlight.Render(m.form.GetString("chain"))
  datadirVal = s.Highlight.Render(m.form.GetString("datadir"))
  fmt.Fprintf(&b, "You've successfully generated a %s configuration file for Bitcoin core on the %s network.\n\n", s.Highlight.Render("bitcoin.conf"), chainVal)
  fmt.Fprintf(&b, "You should copy this file to the data directory: %s\n\n", datadirVal)
  fmt.Fprint(&b, "The configuration file will contain examples of ALL possible settings, and comments to help you make sense of them. Read them carefully before making changes.\n\n")
  fmt.Fprintf(&b, "If you want to start over, you can always generate an example configuration with the %s script in the bitcoin repository.\n\n", s.Highlight.Render("contrib/devtools/gen-bitcoin-conf.sh"))
  fmt.Fprint(&b, "Good luck, anon ;)")
  return s.Status.Margin(0, 1).Padding(1, 2).Width(58).Render(b.String()) + "\n\n"
}

func writeConfig() {
    // Write to bitcoin.conf
    f, err := os.Create("bitcoin.conf")
    if err != nil {
      log.Fatal(err)
    }
    defer f.Close()

    tmpl, err := template.ParseFiles("config.tmpl")
    if err != nil {
      log.Fatal(err)
    }

    cfg := Config{
      DataDir:           datadir,
      Network:           network,
      RPCAuth:           rpcauth,
      RPCPort:           rpcport,
      RPCBind:           rpcbind,
      Server:            boolToInt(server),
      MaxConnections:    maxconnections,
      TxIndex:           boolToInt(txindex),
      Prune:             prune,
      IncludeConf:       includeConf,
      LoadBlock:         loadBlock,
      MaxMempool:        maxMempool,
      MaxOrphanTx:       maxOrphanTx,
      MempoolExpiry:     mempoolExpiry,
      Par:               par,
      PersistMempool:    boolToInt(persistMempool),
      PersistMempoolV1:  boolToInt(persistMempoolV1),
      Pid:               pid,
      Reindex:           boolToInt(reindex),
      ReindexChainstate: boolToInt(reindexChainstate),
      Settings:          settings,
      ShutdownNotify:    shutdownNotify,
      StartupNotify:     startupNotify,
    }

    err = tmpl.Execute(f, cfg)
    if err != nil {
      log.Fatal(err)
    }
}

func main() {
	_, err := tea.NewProgram(NewModel()).Run()
	if err != nil {
		fmt.Println("Oh no:", err)
		os.Exit(1)
	}
  writeConfig()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
