package main

import (
	"log"
	"os"
	
	"github.com/charmbracelet/huh"
)

var (
  datadir  string
)

func main() {

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("datadir").
				Value(&datadir),
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
}