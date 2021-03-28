package main

import (
	"flag"
	"fmt"
	"os"
)

// TODO: Make error for big.Int parsing problem.

var usage = `Usage:
	atto n[ew]
	atto [-a ACCOUNT_INDEX] a[ddress]
	atto [-a ACCOUNT_INDEX] b[alance]
	atto [-a ACCOUNT_INDEX] r[epresentative] REPRESENTATIVE
	atto [-a ACCOUNT_INDEX] s[end] AMOUNT RECEIVER
`

var accountIndexFlag uint

func init() {
	flag.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	flag.UintVar(&accountIndexFlag, "a", 0, "")
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
	return nil // TODO
}

func main() {
	var err error
	switch flag.Arg(0)[:1] {
	case "n":
		err = printNewSeed()
	case "a":
		err = printAddress()
	case "b":
		err = printBalance()
	case "r":
		err = changeRepresentative()
	case "s":
		err = sendFunds()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}
}
