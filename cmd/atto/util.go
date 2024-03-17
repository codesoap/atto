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

// getSeed returns the first line of the standard input.
func getSeed() (string, error) {
	in := bufio.NewReader(os.Stdin)
	firstLine, err := in.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(firstLine), nil
}

func rawToNanoString(raw *big.Int) string {
	rawPerNano, _ := big.NewInt(0).SetString("1000000000000000000000000000000", 10)
	absRaw := big.NewInt(0).Abs(raw)
	integerDigits, fractionalDigits := big.NewInt(0).QuoRem(absRaw, rawPerNano, big.NewInt(0))
	res := integerDigits.String()
	if fractionalDigits.Sign() != 0 {
		fractionalDigitsString := fmt.Sprintf("%030s", fractionalDigits.String())
		res += "." + strings.TrimRight(fractionalDigitsString, "0")
	}
	if raw.Sign() < 0 {
		return "-" + res + " NANO"
	}
	return res + " NANO"
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

func fillWork(block *atto.Block, node string) error {
	switch workSource {
	case workSourceLocal:
		return block.GenerateWork()
	case workSourceNode:
		return block.FetchWork(node)
	case workSourceLocalFallback:
		if err := block.FetchWork(node); err != nil {
			fmt.Fprintf(os.Stderr, "Could not fetch work from node (error: %v); generating it on CPU... ", err)
			return block.GenerateWork()
		}
		return nil
	}
	return fmt.Errorf("unknown work source")
}
