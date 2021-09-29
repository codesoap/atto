package main

import (
	"bufio"
	"fmt"
	"math/big"
	"os"
	"runtime"
	"strings"
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
