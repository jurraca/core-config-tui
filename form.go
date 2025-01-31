package form

import (
	"github.com/charmbracelet/huh"
)

func (m Model) getForm() *huh.Form {
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewNote().Title(m.styles.Note.Render("Basics")),
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
		),
		huh.NewGroup(
			huh.NewConfirm().
				Key("server").
				Title("Enable RPC Server").
				Description("Accept command line and JSON-RPC commands. (default: true)").
				Value(&server)),
		huh.NewGroup(
			huh.NewNote().Title(m.styles.Note.Render("RPC Configuration: the following fields are optional.\nIf you're running Bitcoin Core locally,\nRPC should work as is.")),
			huh.NewInput().
				Key("rpcauth").
				Title("RPC Auth").
				Description("Username and HMAC-SHA-256 hashed password\nfor JSON-RPC connections.\nSee the canonical python script included in\nshare/rpcauth to generate this value.\nDefaults to cookie authentication.").
				Value(&rpcauth),
			huh.NewInput().
				Key("rpcport").
				Title("RPC Port").
				Description("Port for RPC connections (default: 8332)").
				Value(&rpcport),
			huh.NewInput().
				Key("rpcallowip").
				Title("RPC Allow IP").
				Description("Allow JSON-RPC connections from specified source.").
				Value(&rpcallowip),
			huh.NewInput().
				Key("rpcbind").
				Title("RPC bind").
				Description("Bind to given address to listen for JSON-RPC connections.").
				Value(&rpcbind),
		).WithHideFunc(func() bool { return !server }).Title("RPCs"),
		huh.NewGroup(
			huh.NewNote().Title(m.styles.Note.Render("Mempool Options: ")),
			huh.NewInput().
			  Key("maxMempool").
				Title("Max Mempool").
				Description("Keep the transaction memory pool below <n> megabytes (default: 300)").
				Value(&maxMempool),
			huh.NewInput().
			  Key("mempoolExpiry").
				Title("Mempool Expiry").
				Description("Do not keep transactions in the mempool longer than <n> hours (default: 336)").
				Value(&mempoolExpiry),
			huh.NewConfirm().
			  Key("persistMempool").
				Title("Persist Mempool").
				Description("Whether to save the mempool on shutdown and load on restart (default: true)").
				Value(&persistMempool),
		),
		huh.NewGroup(
			huh.NewNote().Title(m.styles.Note.Render("Wallet Options:\nBitcoin Core includes a wallet which is disabled\nby default. If you have an existing wallet,\nyou probably want to keep this disabled.")),
		),
		huh.NewGroup(
			huh.NewNote().Title(m.styles.Note.Render("Danger Zone: leave these blank unless you know what you're doing.")),
		),
	).WithWidth(55).
		WithShowHelp(false).
		WithShowErrors(false)
	
	return m.form
}