
package main

import (
	"log"
	"os"
	"text/template"
	
	"github.com/charmbracelet/huh"
)

type Config struct {
	DataDir        string
	Network        string
	RPCUser        string
	RPCPassword    string
	RPCPort        string
	Server         bool
	MaxConnections string
	TxIndex        bool
	Prune          string
}

var (
	datadir          string
	network          string
	server           bool
	rpcuser          string
	rpcpassword      string
	rpcport          string
	maxconnections   string
	includeConf      string
	loadBlock        string
	maxMempool       string
	maxOrphanTx      string
	mempoolExpiry    string
	par              string
	persistMempool   bool
	persistMempoolV1 bool
	pid              string
	prune            string
	reindex          bool
	reindexChainstate bool
	settings         string
	shutdownNotify   string
	startupNotify    string
	txindex          bool
)

func main() {	
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Bitcoin Data Directory").
				Description("Directory to store blockchain data").
				Value(&datadir),

			huh.NewSelect[string]().
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

			huh.NewInput().
				Title("Include Config").
				Description("Specify additional configuration file, relative to the -datadir path").
				Value(&includeConf),

			huh.NewInput().
				Title("loadblock").
				Description("External filepath to import blocks from on startup").
				Value(&loadBlock),

			huh.NewInput().
				Title("Max Mempool").
				Description("Keep the transaction memory pool below <n> megabytes (default: 300)").
				Value(&maxMempool),

			huh.NewInput().
				Title("Max Orphan Transactions").
				Description("Keep at most <n> unconnectable transactions in memory (default: 100)").
				Value(&maxOrphanTx),

			huh.NewInput().
				Title("Mempool Expiry").
				Description("Do not keep transactions in the mempool longer than <n> hours (default: 336)").
				Value(&mempoolExpiry),

			huh.NewInput().
				Title("Script Verification Threads").
				Description("Set the number of script verification threads (0 = auto, up to 15, <0 = leave that many cores free, default: 0)").
				Value(&par),

			huh.NewConfirm().
				Title("Persist Mempool").
				Description("Whether to save the mempool on shutdown and load on restart (default: 1)").
				Value(&persistMempool),

			huh.NewConfirm().
				Title("Use Legacy Mempool Format").
				Description("Whether a mempool.dat file created by -persistmempool or the savemempool RPC will be written in the legacy format (version 1) or current format (version 2). This temporary option will be removed in the future. (default: 0)").
				Value(&persistMempoolV1),

			huh.NewInput().
				Title("PID File").
				Description("Specify pid file (default: bitcoind.pid)").
				Value(&pid),

			huh.NewInput().
				Title("Prune Blockchain").
				Description("Reduce storage by pruning old blocks (>=550 MB to retain)").
				Value(&prune),

			huh.NewConfirm().
				Title("Reindex").
				Description("Rebuild chain state and block index from blk*.dat files").
				Value(&reindex),

			huh.NewConfirm().
				Title("Reindex Chainstate").
				Description("If enabled, wipe chain state, and rebuild it from blk*.dat files on disk. If an assumeutxo snapshot was loaded, its chainstate will be wiped as well. The snapshot can then be reloaded via RPC.").
				Value(&reindexChainstate),

			huh.NewInput().
				Title("Settings File").
				Description("Path to dynamic settings data file (default: settings.json)").
				Value(&settings),

			huh.NewInput().
				Title("Shutdown Notify Command").
				Description("Execute command immediately before beginning shutdown.").
				Value(&shutdownNotify),

			huh.NewInput().
				Title("Startup Notify Command").
				Description("Execute command on startup").
				Value(&startupNotify),

			huh.NewConfirm().
				Title("Transaction Index").
				Description("Maintain a full transaction index").
				Value(&txindex),
					
			huh.NewInput().
				Title("RPC Username").
				Description("Username for JSON-RPC connections").
				Value(&rpcuser),

			huh.NewInput().
				Title("RPC Password").
				Description("Password for JSON-RPC connections").
				Value(&rpcpassword),

			huh.NewInput().
				Title("RPC Port").
				Description("Port for RPC connections (default: 8332)").
				Value(&rpcport),

			huh.NewInput().
				Title("Max Connections").
				Description("Max peer connections (default: 125)").
				Value(&maxconnections),

			huh.NewInput().
				Title("Prune Blockchain").
				Description("Reduce storage (0 for no pruning, >=550 for MB to retain)").
				Value(&prune),
		).Title("General Options"),
	)

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

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
		DataDir:        datadir,
		Network:        network,
		RPCUser:        rpcuser,
		RPCPassword:    rpcpassword,
		RPCPort:        rpcport,
		Server:         server,
		MaxConnections: maxconnections,
		TxIndex:        txindex,
		Prune:          prune,
	}

	err = tmpl.Execute(f, cfg)
	if err != nil {
		log.Fatal(err)
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}