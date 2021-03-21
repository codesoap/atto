package main

import (
	"flag"
	"fmt"
	"os"
)

var usage = `Usage:
	atto n[ew]
	atto [--account-index ACCOUNT_INDEX] r[epresentative] REPRESENTATIVE
	atto [(-n COUNT|-a ACCOUNT_INDEX)] a[ddress]
	atto [--account-index ACCOUNT_INDEX] [--no-receive] b[alance]
	atto [--account-index ACCOUNT_INDEX] [--no-confirm] s[end] RECEIVER
`

var accountIndexFlag, countFlag uint
var noReceiveFlag, noConfirmFlag bool

func init() {
	flag.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	flag.UintVar(&accountIndexFlag, "a", 0, "")
	flag.UintVar(&accountIndexFlag, "account-index", 0, "")
	flag.UintVar(&countFlag, "n", 0, "")
	flag.BoolVar(&noReceiveFlag, "no-receive", false, "")
	flag.BoolVar(&noConfirmFlag, "no-confirm", false, "")
	flag.Parse()
	if err := verifyLegalUsage(); err != nil {
		flag.Usage()
		os.Exit(1)
	}
}

func verifyLegalUsage() error {
	// Flags match command.
	// flags.NArgs OK.
	// Subcommand exists.
	// Index positive
	// Count positive
	return nil // TODO
}

func main() {
	var err error
	switch flag.Arg(0)[:1] {
	case "n":
		err = printNewSeed()
	case "a":
		err = printAddresses()
	case "r":
	case "b":
		err = printBalance()
	case "s":
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}
}
