
package main

import (
	"log"
	"os"
	"text/template"
	
	"github.com/charmbracelet/huh"
)

var (
	datadir     string
	network     string
	rpcuser     string
	rpcpassword string
	rpcport     string
	maxconnections string
	server      bool
	txindex     bool
	prune       string
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
					huh.NewOption("mainnet", "mainnet"),
					huh.NewOption("testnet", "testnet"),
					huh.NewOption("regtest", "regtest"),
					huh.NewOption("signet", "signet"),
				).
			  Value(&network),

			huh.NewConfirm().
				Title("Enable Server?").
				Description("Accept command-line and JSON-RPC commands").
				Value(&server),
					
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

			huh.NewConfirm().
				Title("Enable Transaction Index?").
				Description("Maintain a full transaction index").
				Value(&txindex),

			huh.NewInput().
				Title("Prune Blockchain").
				Description("Reduce storage (0 for no pruning, >=550 for MB to retain)").
				Value(&prune),
		),
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

	tmpl, err := template.ParseFiles("config.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	cfg := Config{
		DataDir:        datadir,
		Network:        handleNetwork(network),
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

	_, err = f.WriteString(config)
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

func handleNetwork(network string) string {
	switch network {
	case "mainnet":
		return "chain=mainnet"
	case "testnet":
		return "chain=testnet"
	case "regtest":
		return "chain=regtest"
	case "signet":
		return "chain=testnet\nsignet=1"
	default:
		return ""
	}
}