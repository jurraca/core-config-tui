
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	
	"github.com/charmbracelet/huh"
)

var (
	datadir     string
	testnet     bool
	rpcuser     string
	rpcpassword string
	rpcport     string
	maxconnections string
	server      bool
	txindex     bool
	prune       string
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	defaultDataDir := filepath.Join(homeDir, ".bitcoin")

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Bitcoin Data Directory").
				Description("Directory to store blockchain data").
				Value(&datadir).
				Default(defaultDataDir),

			huh.NewConfirm().
				Title("Enable Testnet?").
				Description("Run on Bitcoin's test network").
				Value(&testnet),

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
				Value(&rpcport).
				Default("8332"),

			huh.NewInput().
				Title("Max Connections").
				Description("Max peer connections (default: 125)").
				Value(&maxconnections).
				Default("125"),

			huh.NewConfirm().
				Title("Enable Server?").
				Description("Accept command-line and JSON-RPC commands").
				Value(&server),

			huh.NewConfirm().
				Title("Enable Transaction Index?").
				Description("Maintain a full transaction index").
				Value(&txindex),

			huh.NewInput().
				Title("Prune Blockchain").
				Description("Reduce storage (0 for no pruning, >=550 for MB to retain)").
				Value(&prune).
				Default("0"),
		),
	)

	err = form.Run()
	if err != nil {
		log.Fatal(err)
	}

	// Write to bitcoin.conf
	f, err := os.Create("bitcoin.conf")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	config := fmt.Sprintf(`# Bitcoin Core configuration file
# Generated by Bitcoin Core TUI

# Data directory
datadir=%s

# Network
testnet=%v

# RPC Settings
rpcuser=%s
rpcpassword=%s
rpcport=%s
server=%v

# Network Settings
maxconnections=%s

# Index Settings
txindex=%v

# Storage Settings
prune=%s
`, datadir, boolToInt(testnet), rpcuser, rpcpassword, rpcport, 
   boolToInt(server), maxconnections, boolToInt(txindex), prune)

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
