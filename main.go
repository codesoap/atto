package main

import (
	"flag"
	"fmt"
	"os"
)

var usage = `Usage:
	atto -v
	atto n[ew]
	atto [-a ACCOUNT_INDEX] a[ddress]
	atto [-a ACCOUNT_INDEX] b[alance]
	atto [-a ACCOUNT_INDEX] r[epresentative] REPRESENTATIVE
	atto [-a ACCOUNT_INDEX] [-y] s[end] AMOUNT RECEIVER

If the -v flag is provided, atto will print its version number.

The new subcommand generates a new seed, which can later be used with
the other subcommands.

The address, balance, representative and send subcommands expect a seed
as as the first line of their standard input. Showing the first address
of a newly generated key could work like this:
atto new | tee seed.txt | atto address

The send subcommand also expects manual confirmation of the transaction,
unless the -y flag is given.

The address subcommand displays addresses for a seed, the balance
subcommand receives pending sends and shows the balance of an account,
the representative subcommand changes the account's representative and
the send subcommand sends funds to an address.

ACCOUNT_INDEX is an optional parameter, which must be a number between 0
and 4,294,967,295. It allows you to use multiple accounts derived from
the same seed. By default the account with index 0 is chosen.
`

var accountIndexFlag uint
var yFlag bool

func init() {
	var vFlag bool
	flag.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	flag.UintVar(&accountIndexFlag, "a", 0, "")
	flag.BoolVar(&yFlag, "y", false, "")
	flag.BoolVar(&vFlag, "v", false, "")
	flag.Parse()
	if vFlag {
		fmt.Println("1.2.0")
		os.Exit(0)
	}
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	var ok bool
	switch flag.Arg(0)[:1] {
	case "n", "a", "b":
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
