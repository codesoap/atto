package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"

	"github.com/codesoap/atto"
)

var usage = `Usage:
	atto-safesign -v
	atto-safesign FILE receive
	atto-safesign FILE representative REPRESENTATIVE
	atto-safesign FILE send AMOUNT RECEIVER
	atto-safesign [-a ACCOUNT_INDEX] [-y] FILE sign
	atto-safesign FILE submit

If the -v flag is provided, atto-safesign will print its version number.

The receive, representative, send and submit subcommands expect a Nano
address as the first line of their standard input. This address will be
the account of the generated and submitted blocks.

The receive, representative and send subcommands will generate blocks
and append them to FILE. The blocks will still be lacking their
signature. The receive subcommand will create multiple blocks, if there
are multiple pending sends that can be received. The representative
subcommand will create a block for changing the representative and the
send subcommand will create a block for sending funds to an address.

The sign subcommand expects a seed as the first line of standard input.
It also expects manual confirmation before signing blocks, unless the
-y flag is given. The seed and ACCOUNT_INDEX must belong to the address
used when creating blocks with receive, representative or send.

The sign subcommand will add signatures to all blocks in FILE. It is the
only subcommand that requires no network connection.

The submit subcommand will submit all blocks contained in FILE to the
Nano network.

ACCOUNT_INDEX is an optional parameter, which allows you to use
different accounts derived from the given seed. By default the account
with index 0 is chosen.

Environment:
	ATTO_BASIC_AUTH_USERNAME  The username for HTTP Basic Authentication.
	                          If set, HTTP Basic Authentication will be
	                          used when making requests to the node.
	ATTO_BASIC_AUTH_PASSWORD  The password to use for HTTP Basic
	                          Authentication.
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
	if accountIndexFlag >= 1<<32 || flag.NArg() < 2 {
		flag.Usage()
		os.Exit(1)
	}
	var ok bool
	switch flag.Arg(1) {
	case "receive", "sign", "submit":
		ok = flag.NArg() == 2
	case "representative":
		ok = flag.NArg() == 3
	case "send":
		ok = flag.NArg() == 4
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
	switch flag.Arg(1) {
	case "receive":
		err = receive()
	case "representative":
		err = change()
	case "send":
		err = send()
	case "sign":
		err = sign()
	case "submit":
		err = submit()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}
}

func receive() error {
	addr, err := getFirstStdinLine()
	if err != nil {
		return err
	}
	account, err := atto.NewAccountFromAddress(addr)
	if err != nil {
		return err
	}
	firstReceive := false // Is this the very first block of the account?
	info, err := getLatestAccountInfo(account)
	if err == atto.ErrAccountNotFound {
		firstReceive = true
	} else if err != nil {
		return err
	}
	pendings, err := account.FetchPending(node)
	if err != nil {
		return err
	}
	for _, pending := range pendings {
		var block atto.Block
		if firstReceive {
			info, block, err = account.FirstReceive(pending, defaultRepresentative)
			firstReceive = false
		} else {
			block, err = info.Receive(pending)
		}
		if err != nil {
			return err
		}
		if err = block.FetchWork(node); err != nil {
			return err
		}
		blockJSON, err := json.Marshal(block)
		if err != nil {
			return err
		}
		err = appendLineToFile(blockJSON)
		if err != nil {
			return err
		}
	}
	return nil
}

func change() error {
	representative := flag.Arg(2)
	addr, err := getFirstStdinLine()
	if err != nil {
		return err
	}
	account, err := atto.NewAccountFromAddress(addr)
	if err != nil {
		return err
	}
	info, err := getLatestAccountInfo(account)
	if err != nil {
		return err
	}
	block, err := info.Change(representative)
	if err != nil {
		return err
	}
	if err = block.FetchWork(node); err != nil {
		return err
	}
	blockJSON, err := json.Marshal(block)
	if err != nil {
		return err
	}
	return appendLineToFile(blockJSON)
}

func send() error {
	amount := flag.Arg(2)
	receiver := flag.Arg(3)
	addr, err := getFirstStdinLine()
	if err != nil {
		return err
	}
	account, err := atto.NewAccountFromAddress(addr)
	if err != nil {
		return err
	}
	info, err := getLatestAccountInfo(account)
	if err != nil {
		return err
	}
	block, err := info.Send(amount, receiver)
	if err != nil {
		return err
	}
	if err = block.FetchWork(node); err != nil {
		return err
	}
	blockJSON, err := json.Marshal(block)
	if err != nil {
		return err
	}
	return appendLineToFile(blockJSON)
}

func sign() error {
	seed, err := getFirstStdinLine()
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
	blocks, err := getBlocksFromFile()
	if err != nil {
		return err
	}
	var outBuffer bytes.Buffer
	for _, block := range blocks {
		if account.Address != block.Account {
			txt := "Used account with address '%s' cannot sign block with address '%s'"
			return fmt.Errorf(txt, account.Address, block.Account)
		}
		if err = letUserVerifyBlock(block); err != nil {
			return err
		}
		block.Sign(privateKey)
		blockJSON, err := json.Marshal(block)
		if err != nil {
			return err
		}

		// Buffer output so that file can be overwritten as late as possible
		// to avoid problems during the write as much as possible.
		outBuffer.Write(blockJSON)    // err is always nil.
		outBuffer.Write([]byte{'\n'}) // err is always nil.
	}

	file, err := os.Create(flag.Arg(0))
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, &outBuffer)
	return err
}

func submit() error {
	addr, err := getFirstStdinLine()
	if err != nil {
		return err
	}
	account, err := atto.NewAccountFromAddress(addr)
	if err != nil {
		return err
	}
	blocks, err := getBlocksFromFile()
	if err != nil {
		return err
	}

	var oldBalance *big.Int
	info, err := account.FetchAccountInfo(node)
	if err == atto.ErrAccountNotFound {
		oldBalance = big.NewInt(0)
	} else if err != nil {
		return err
	} else {
		var ok bool
		oldBalance, ok = big.NewInt(0).SetString(info.Balance, 10)
		if !ok {
			return fmt.Errorf("cannot parse '%s' as an integer", info.Balance)
		}
	}

	for _, block := range blocks {
		newBalance, ok := big.NewInt(0).SetString(block.Balance, 10)
		if !ok {
			return fmt.Errorf("cannot parse '%s' as an integer", block.Balance)
		}
		switch oldBalance.Cmp(newBalance) {
		case -1:
			block.SubType = atto.SubTypeReceive
		case 0:
			// If the balance does not change, this should be a "change" block.
			block.SubType = atto.SubTypeChange
		case 1:
			block.SubType = atto.SubTypeSend
		}
		fmt.Fprint(os.Stderr, "Submitting block... ")
		err = block.Submit(node)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "done")
		oldBalance = newBalance
	}
	return nil
}
