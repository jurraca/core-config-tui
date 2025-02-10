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
	Server            bool
	MaxConnections    string
	TxIndex           bool
	Prune             string
	IncludeConf       string
	LoadBlock         string
	MaxMempool        string
	MaxOrphanTx       string
	MempoolExpiry     string
	Par               string
	PersistMempool    bool
	PersistMempoolV1  bool
	Pid               string
	Reindex           bool
	ReindexChainstate bool
	Settings          string
	ShutdownNotify    string
	StartupNotify     string
	DisableWallet     bool
	Wallet            string
	WalletDir         string
	WalletRBF         bool
}

var (
	cfg = Config{
		Server:         true,
		PersistMempool: true,
		DisableWallet:  true,
		WalletRBF:      true,
	}

	red       = lipgloss.AdaptiveColor{Light: "#FE5F86", Dark: "#FE5F86"}
	indigo    = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
	green     = lipgloss.AdaptiveColor{Light: "#02BA84", Dark: "#02BF87"}
	orange    = lipgloss.AdaptiveColor{Light: "#FFA500", Dark: "#FF8C00"}
	highlight = lipgloss.AdaptiveColor{Light: "#FFA500", Dark: "#FF8C00"}

	maybeReindex = false
	confirmReindex = false
)

type Styles struct {
	Base,
	HeaderText,
	Status,
	StatusHeader,
	Note,
	Highlight,
	ErrorHeaderText,
	Warning,
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
		Foreground(orange).
		Bold(true)
	s.Note = lg.NewStyle().
		Background(highlight).
		Foreground(lipgloss.Color("#FAFAFA")).
		Bold(true).
		MarginLeft(1).
		MarginTop(1).
		Padding(1, 2, 1, 2)
	s.Highlight = lg.NewStyle().
		Foreground(lipgloss.Color("208"))
	s.ErrorHeaderText = s.HeaderText.
		Foreground(red)
	s.Warning = lg.NewStyle().
		Background(lipgloss.Color("196")).
		Foreground(lipgloss.Color("#FAFAFA")).
		Bold(true).
		MarginLeft(1).
		MarginTop(1).
		Padding(1, 2, 1, 2)
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
	height int
}

func NewModel() Model {
	m := Model{width: maxWidth}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewNote().Title(m.styles.Note.Render("Basics")),
			huh.NewInput().
				Key("datadir").
				Title("Data Directory").
				Description("Directory to store blockchain data\n(defaults to ~/.bitcoin on Linux)").
				Value(&cfg.DataDir),

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
				Value(&cfg.Network),

			huh.NewConfirm().
				Key("txindex").
				Title("Transaction Index").
				Description("Maintain a full transaction index, used by the\ngetrawtransaction rpc call (default: No)").
				Value(&cfg.TxIndex),
		),
		huh.NewGroup(
			huh.NewNote().Title(m.styles.Note.Render("Prune Blockchain\nSince you are not storing a txIndex,\nyou can limit the size of data you keep on disk.")),
			huh.NewInput().
				Key("prune").
				Title("Prune").
				Description("Prune the blockchain database.\nPossible values: \n 0 = disable pruning blocks (default),\n 1 = allow manual pruning via RPC,\n >=550 = automatically prune block files\n to stay under the specified target size in MiB").
				Validate(func(v string) error {
					if cfg.TxIndex == true && v != "0" {
						return fmt.Errorf("pruning is incompatible with txindex. If you want to use pruning, you must disable txindex.")
					}
					return nil
				}).
				Value(&cfg.Prune),
		).WithHideFunc(func() bool { return cfg.TxIndex }),
		huh.NewGroup(
			huh.NewNote().Title(m.styles.Note.Render("RPC Configuration")),
			huh.NewConfirm().
				Key("server").
				Title("Enable RPC Server").
				Description("Accept command line and JSON-RPC commands.\n(default: Yes)").
				Value(&cfg.Server)),
		huh.NewGroup(
			huh.NewNote().Title(m.styles.Note.Render("RPC Configuration\nThe following fields are optional.\nIf you're running Bitcoin Core locally,\nRPC should work as is.")),
			huh.NewInput().
				Key("rpcauth").
				Title("RPC Auth").
				Description("Username and HMAC-SHA-256 hashed password\nfor JSON-RPC connections.\nSee the canonical python script included in\nshare/rpcauth to generate this value.\nDefaults to cookie authentication.").
				Value(&cfg.RPCAuth),
			huh.NewInput().
				Key("rpcport").
				Title("RPC Port").
				Description("Port for RPC connections (default: 8332)").
				Value(&cfg.RPCPort),
			huh.NewInput().
				Key("rpcallowip").
				Title("RPC Allow IP").
				Description("Allow JSON-RPC connections from specified source.").
				Value(&cfg.RPCAllowIP),
			huh.NewInput().
				Key("rpcbind").
				Title("RPC Bind").
				Description("Bind to given address to listen for JSON-RPC\nconnections.").
				Value(&cfg.RPCBind),
		).WithHideFunc(func() bool { return !cfg.Server }).Title("RPCs"),
		huh.NewGroup(
			huh.NewNote().Title(m.styles.Note.Render("Mempool Options")),
			huh.NewInput().
				Key("maxMempool").
				Title("Max Mempool").
				Description("Keep the transaction memory pool below <n> megabytes\n(default: 300)").
				Value(&cfg.MaxMempool),
			huh.NewInput().
				Key("mempoolExpiry").
				Title("Mempool Expiry").
				Description("Do not keep transactions in the mempool longer than\n<n> hours (default: 336)").
				Value(&cfg.MempoolExpiry),
			huh.NewConfirm().
				Key("persistMempool").
				Title("Persist Mempool").
				Description("Whether to save the mempool on shutdown\nand load on restart (default: Yes)").
				Value(&cfg.PersistMempool),
		),
		huh.NewGroup(
			huh.NewNote().Title(m.styles.Note.Render("Disable Wallet:\nBitcoin Core includes a wallet which is disabled\nby default. If you have an existing wallet,\nyou probably want to keep this disabled.")),
			huh.NewConfirm().
				Key("disablewallet").
				Title("Disable Wallet").
				Description("Disable the wallet and wallet RPC calls\n(default: Yes)").
				Value(&cfg.DisableWallet),
		),
		huh.NewGroup(
			huh.NewNote().Title(m.styles.Note.Render("Wallet Options: ")),
			huh.NewInput().
				Key("wallet").
				Title("Wallet Path").
				Description("Specify wallet path to load at startup.\nCan be used multiple times to load multiple wallets.").
				Value(&cfg.Wallet),
			huh.NewInput().
				Key("walletdir").
				Title("Wallet Directory").
				Description("Directory to hold wallets (default: <datadir>/wallets\nif it exists, otherwise <datadir>)").
				Value(&cfg.WalletDir),
			huh.NewConfirm().
				Key("walletrbf").
				Title("Wallet RBF").
				Description("Send transactions with full-RBF opt-in\nenabled (default: true)").
				Value(&cfg.WalletRBF),
		).WithHideFunc(func() bool { return cfg.DisableWallet }),
		huh.NewGroup(
			huh.NewNote().Title(m.styles.Note.Render("Reindex Blockchain Data")),
			huh.NewConfirm().
				Title("Do you want to download the blockchain from genesis?").
				Value(&maybeReindex),
		),
		huh.NewGroup(
			huh.NewNote().Title(m.styles.Warning.Render("Danger Zone - Reindex Blockchain:\n")),
			huh.NewConfirm().
			        Title("Are you sure?").
				Description("Enabling the following settings will ERASE existing\nchain and/or index data. Downloading the chain\nfrom genesis and reindexing is a time-consuming\nprocess.").
				Affirmative("I understand.").
				Negative("Nope!").
				Value(&confirmReindex),
			huh.NewConfirm().
				Key("reindexChainstate").
				Title("Reindex Chainstate").
				Description("Wipe chain state and block index, and rebuild them\nfrom blk*.dat files on disk.").
				Validate(func(v bool) error {
					if confirmReindex == false && v == true {
						return fmt.Errorf("Please confirm above before enabling this setting.")
					}
					return nil
				}).
				Value(&cfg.Reindex),
			huh.NewConfirm().
				Key("reindex").
				Title("Reindex").
				Description("Wipe chain state and block index, and rebuild them\nfrom blk*.dat files on disk. Also wipe and rebuild\nother optional indexes that are active.").
				Validate(func(v bool) error {
					if confirmReindex == false && v == true {
						return fmt.Errorf("Please confirm above before enabling this setting.")
					}
					return nil
				}).
				Value(&cfg.ReindexChainstate),
		).WithHideFunc(func() bool { return !maybeReindex }),
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
		header := m.appBoundaryView("Configuration Completed")
		body := m.CompletedMsg(*s)
		return s.Base.Render(header + "\n" + body + "\n")

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
		datadir           string
		chain             string
		txindex           string
		prune             string
		server            string
		reindex           string
		reindexChainstate string
		rpcauth           string
		rpcport           string
		rpcallowip        string
		rpcbind           string
		maxMempool        string
		mempoolExpiry     string
		persistMempool    string
		disableWallet     string
		wallet            string
		walletdir         string
		walletrbf         string
	)
	if m.form.GetString("datadir") != "" {
		datadir = "datadir: " + m.form.GetString("datadir") + "\n"
	}
	if m.form.GetString("chain") != "" {
		chain = "chain: " + m.form.GetString("chain") + "\n"
	}
	if m.form.GetBool("txindex") != false {
		txindex = "txindex: " + strconv.FormatBool(m.form.GetBool("txindex")) + "\n"
	}
	if m.form.GetBool("server") == true {
		server = "server: " + strconv.FormatBool(m.form.GetBool("server")) + "\n"
	}
	if m.form.GetString("prune") != "" {
		prune = "prune: " + m.form.GetString("prune") + " MiB" + "\n"
	}
	if m.form.GetString("rpcauth") != "" {
		rpcauth = "rpcauth: " + m.form.GetString("rpcauth") + "\n"
	}
	if m.form.GetString("rpcport") != "" {
		rpcport = "rpcport: " + m.form.GetString("rpcport") + "\n"
	}
	if m.form.GetString("rpcallowip") != "" {
		rpcallowip = "allowip: " + m.form.GetString("rpcallowip") + "\n"
	}
	if m.form.GetString("rpcbind") != "" {
		rpcbind = "rpcbind: " + m.form.GetString("rpcbind") + "\n"
	}
	if m.form.GetString("maxMempool") != "" {
		maxMempool = "maxmempool: " + m.form.GetString("maxMempool") + "\n"
	}
	if m.form.GetString("mempoolExpiry") != "" {
		mempoolExpiry = "mempoolexpiry: " + m.form.GetString("mempoolExpiry") + "\n"
	}
	if m.form.GetBool("persistMempool") != false {
		persistMempool = "persistmempool: " + strconv.FormatBool(m.form.GetBool("persistMempool")) + "\n"
	}
	if m.form.GetBool("disablewallet") != false {
		disableWallet = "disableWallet: " + strconv.FormatBool(m.form.GetBool("disablewallet")) + "\n"
	}
	if m.form.GetString("Wallet") != "" {
		wallet = "wallet: " + m.form.GetString("wallet") + "\n"
	}
	if m.form.GetString("walletdir") != "" {
		walletdir = "walletdir: " + m.form.GetString("walletdir") + "\n"
	}
	if m.form.GetBool("walletrbf") != false {
		walletrbf = "walletrbf: " + strconv.FormatBool(m.form.GetBool("walletrbf")) + "\n"
	}
	if m.form.GetBool("reindex") != false {
		reindex = "reindex: " + strconv.FormatBool(m.form.GetBool("reindex")) + "\n"
	}
	if m.form.GetBool("reindexChainstate") != false {
		reindexChainstate = "reindexChainstate: " + strconv.FormatBool(m.form.GetBool("reindexChainstate")) + "\n"
	}
	const statusWidth = 32
	statusMarginLeft := m.width - statusWidth - lipgloss.Width(form.View()) - s.Status.GetMarginRight() - 2
	return s.Status.
		Height(max(lipgloss.Height(form.View()), 28)).
		Width(statusWidth).
		MarginLeft(statusMarginLeft).
		Render(s.StatusHeader.Render("Current Config") +
			"\n\n" +
			datadir +
			chain +
			txindex +
			prune +
			server +
			rpcauth +
			rpcport +
			rpcallowip +
			rpcbind +
			maxMempool +
			mempoolExpiry +
			persistMempool +
			disableWallet +
			wallet +
			walletdir +
			walletrbf +
			reindex +
			reindexChainstate + "\n",
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
	return s.Status.Margin(1, 4).Padding(1, 2).Width(58).Render(b.String()) + "\n"
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

	type ConfigTemplate struct {
		Config
		Server            int
		TxIndex           int
		PersistMempool    int
		PersistMempoolV1  int
		Reindex           int
		ReindexChainstate int
		DisableWallet     int
		WalletRBF         int
	}

	templateData := ConfigTemplate{
		Config:            cfg,
		Server:            boolToInt(cfg.Server),
		TxIndex:           boolToInt(cfg.TxIndex),
		PersistMempool:    boolToInt(cfg.PersistMempool),
		PersistMempoolV1:  boolToInt(cfg.PersistMempoolV1),
		Reindex:           boolToInt(cfg.Reindex),
		ReindexChainstate: boolToInt(cfg.ReindexChainstate),
		DisableWallet:     boolToInt(cfg.DisableWallet),
		WalletRBF:         boolToInt(cfg.WalletRBF),
	}

	// Convert boolean fields to int
	templateData.Server = boolToInt(cfg.Server)
	templateData.TxIndex = boolToInt(cfg.TxIndex)
	templateData.PersistMempool = boolToInt(cfg.PersistMempool)
	templateData.PersistMempoolV1 = boolToInt(cfg.PersistMempoolV1)
	templateData.Reindex = boolToInt(cfg.Reindex)
	templateData.ReindexChainstate = boolToInt(cfg.ReindexChainstate)
	templateData.DisableWallet = boolToInt(cfg.DisableWallet)
	templateData.WalletRBF = boolToInt(cfg.WalletRBF)

	err = tmpl.Execute(f, templateData)
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
