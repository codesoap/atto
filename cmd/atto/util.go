package main

import (
	"bufio"
	"fmt"
	"math/big"
	"os"
	"runtime"
	"strings"

	"github.com/codesoap/atto"
)

// ownAccount initializes the own account using the seed provided via
// standard input and accountIndexFlag.
func ownAccount() (atto.Account, error) {
	seed, err := getSeed()
	if err != nil {
		return atto.Account{}, err
	}
	return atto.NewAccount(seed, uint32(accountIndexFlag))
}

// getSeed takes the first line of the standard input and interprets it
// as a hexadecimal representation of a 32byte seed.
func getSeed() (*big.Int, error) {
	in := bufio.NewReader(os.Stdin)
	firstLine, err := in.ReadString('\n')
	if err != nil {
		return nil, err
	}
	seed, ok := big.NewInt(0).SetString(strings.TrimSpace(firstLine), 16)
	if !ok {
		return nil, fmt.Errorf("could not parse seed")
	}
	return seed, nil
}

func rawToNanoString(raw *big.Int) string {
	rawPerKnano, _ := big.NewInt(0).SetString("1000000000000000000000000000", 10)
	balance := big.NewInt(0).Div(raw, rawPerKnano).Uint64()
	return fmt.Sprintf("%d.%03d NANO", balance/1000, balance%1000)
}

func letUserVerifySend(amount, recipient string) (err error) {
	if !yFlag {
		fmt.Printf("Send %s NANO to %s? [y/N]: ", amount, recipient)

		// Explicitly openning /dev/tty or CONIN$ ensures function, even if
		// the standard input is not a terminal.
		var tty *os.File
		if runtime.GOOS == "windows" {
			tty, err = os.Open("CONIN$")
		} else {
			tty, err = os.Open("/dev/tty")
		}
		if err != nil {
			msg := "could not open terminal for confirmation input: %v"
			return fmt.Errorf(msg, err)
		}
		defer tty.Close()

		var confirmation string
		fmt.Fscanln(tty, &confirmation)
		if confirmation != "y" && confirmation != "Y" {
			fmt.Fprintln(os.Stderr, "Send aborted.")
			os.Exit(0)
		}
	}
	return
}
