package main

import (
	"flag"
	"fmt"
	"os"
)

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
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	var ok bool
	switch flag.Arg(0)[:1] {
	case "n":
		ok = flag.NArg() == 1
	case "a":
		ok = flag.NArg() == 1
	case "b":
		ok = flag.NArg() == 1
	case "r":
		ok = flag.NArg() == 2
	case "s":
		ok = flag.NArg() == 3
	}
	if !ok {
		flag.Usage()
		os.Exit(1)
	}
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
