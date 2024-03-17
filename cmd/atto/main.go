package main

import (
	"flag"
	"fmt"
	"math/big"
	"net/http"
	"os"

	"github.com/codesoap/atto"
)

var usage = `Usage:
	atto -v
	atto n[ew]
	atto [-a ACCOUNT_INDEX] a[ddress]
	atto [-a ACCOUNT_INDEX] b[alance]
	atto [-a ACCOUNT_INDEX] r[epresentative] [NEW_REPRESENTATIVE]
	atto [-a ACCOUNT_INDEX] [-y] s[end] AMOUNT RECEIVER

If the -v flag is provided, atto will print its version number.

The new subcommand generates a new seed, which can later be used with
the other subcommands.

The address, balance, representative and send subcommands expect a seed
as the first line of their standard input. Showing the first address of
a newly generated key could work like this:
atto new | tee seed.txt | atto address

The send subcommand also expects manual confirmation of the transaction,
unless the -y flag is given.

The address subcommand displays addresses for a seed, the balance
subcommand receives receivable blocks and shows the balance of an
account, the representative subcommand shows the current representative
if NEW_REPRESENTATIVE is not given and changes the account's
representative if it is given and the send subcommand sends funds to an
address.

ACCOUNT_INDEX is an optional parameter, which must be a number between 0
and 4,294,967,295. It allows you to use multiple accounts derived from
the same seed. By default the account with index 0 is chosen.

Environment:
	ATTO_BASIC_AUTH_USERNAME  The username for HTTP Basic Authentication.
	                          If set, HTTP Basic Authentication will be
	                          used when making requests to the node.
	ATTO_BASIC_AUTH_PASSWORD  The password to use for HTTP Basic
	                          Authentication.
`

type workSourceType int

const (
	workSourceLocal workSourceType = iota
	workSourceNode
	workSourceLocalFallback
)

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
		fmt.Println("1.5.0")
		os.Exit(0)
	}
	if accountIndexFlag >= 1<<32 || flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	var ok bool
	switch flag.Arg(0)[:1] {
	case "n", "a", "b":
		ok = flag.NArg() == 1
	case "r":
		ok = flag.NArg() == 1 || flag.NArg() == 2
	case "s":
		ok = flag.NArg() == 3
	}
	if !ok {
		flag.Usage()
		os.Exit(1)
	}
	setUpNodeAuthentication()
}

func setUpNodeAuthentication() {
	if os.Getenv("ATTO_BASIC_AUTH_USERNAME") != "" {
		username := os.Getenv("ATTO_BASIC_AUTH_USERNAME")
		password := os.Getenv("ATTO_BASIC_AUTH_PASSWORD")
		atto.RequestInterceptor = func(request *http.Request) error {
			request.SetBasicAuth(username, password)
			return nil
		}
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
		if flag.NArg() == 1 {
			err = printRepresentative()
		} else {
			err = changeRepresentative()
		}
	case "s":
		err = sendFunds()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}
}

func printNewSeed() error {
	seed, err := atto.GenerateSeed()
	if err == nil {
		fmt.Println(seed)
	}
	return err
}

func printAddress() error {
	seed, err := getSeed()
	if err != nil {
		return err
	}
	privateKey, err := atto.NewPrivateKey(seed, uint32(accountIndexFlag))
	if err != nil {
		return err
	}
	account, err := atto.NewAccount(privateKey)
	if err == nil {
		fmt.Println(account.Address)
	}
	return err
}

func printBalance() error {
	seed, err := getSeed()
	if err != nil {
		return err
	}
	privateKey, err := atto.NewPrivateKey(seed, uint32(accountIndexFlag))
	if err != nil {
		return err
	}
	account, err := atto.NewAccount(privateKey)
	if err != nil {
		return err
	}
	firstReceive := false // Is this the very first block of the account?
	info, err := account.FetchAccountInfo(node)
	if err == atto.ErrAccountNotFound {
		// Needed for printing balance, even if there are no receivable blocks:
		info.Balance = "0"

		firstReceive = true
	} else if err != nil {
		return err
	}
	receivables, err := account.FetchReceivable(node)
	if err != nil {
		return err
	}
	for _, receivable := range receivables {
		txt := "Creating receive block for %s from %s... "
		amount, ok := big.NewInt(0).SetString(receivable.Amount, 10)
		if !ok {
			return fmt.Errorf("cannot parse '%s' as an integer", receivable.Amount)
		}
		fmt.Fprintf(os.Stderr, txt, rawToNanoString(amount), receivable.Source)
		var block atto.Block
		if firstReceive {
			fmt.Fprintf(os.Stderr, "opening account... ")
			info, block, err = account.FirstReceive(receivable, defaultRepresentative)
			firstReceive = false
		} else {
			block, err = info.Receive(receivable)
		}
		if err != nil {
			return err
		}
		if err = block.Sign(privateKey); err != nil {
			return err
		}
		if err = fillWork(&block, node); err != nil {
			return err
		}
		if err = block.Submit(node); err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "done")
	}
	newBalance, ok := big.NewInt(0).SetString(info.Balance, 10)
	if !ok {
		return fmt.Errorf("cannot parse '%s' as an integer", info.Balance)
	}
	fmt.Println(rawToNanoString(newBalance))
	return nil
}

func printRepresentative() error {
	seed, err := getSeed()
	if err != nil {
		return err
	}
	privateKey, err := atto.NewPrivateKey(seed, uint32(accountIndexFlag))
	if err != nil {
		return err
	}
	account, err := atto.NewAccount(privateKey)
	if err != nil {
		return err
	}
	info, err := account.FetchAccountInfo(node)
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, info.Representative)
	return nil
}

func changeRepresentative() error {
	representative := flag.Arg(1)
	seed, err := getSeed()
	if err != nil {
		return err
	}
	privateKey, err := atto.NewPrivateKey(seed, uint32(accountIndexFlag))
	if err != nil {
		return err
	}
	account, err := atto.NewAccount(privateKey)
	if err != nil {
		return err
	}
	info, err := account.FetchAccountInfo(node)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Creating change block... ")
	block, err := info.Change(representative)
	if err != nil {
		return err
	}
	if err = block.Sign(privateKey); err != nil {
		return err
	}
	if err = fillWork(&block, node); err != nil {
		return err
	}
	if err = block.Submit(node); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "done")
	return nil
}

func sendFunds() error {
	amount := flag.Arg(1)
	recipient := flag.Arg(2)
	seed, err := getSeed()
	if err != nil {
		return err
	}
	privateKey, err := atto.NewPrivateKey(seed, uint32(accountIndexFlag))
	if err != nil {
		return err
	}
	account, err := atto.NewAccount(privateKey)
	if err != nil {
		return err
	}
	if err = letUserVerifySend(amount, recipient); err != nil {
		return err
	}
	info, err := account.FetchAccountInfo(node)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Creating send block... ")
	block, err := info.Send(amount, recipient)
	if err != nil {
		return err
	}
	if err = block.Sign(privateKey); err != nil {
		return err
	}
	if err = fillWork(&block, node); err != nil {
		return err
	}
	if err = block.Submit(node); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "done")
	return nil
}
