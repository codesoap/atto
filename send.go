package main

import (
	"flag"
	"fmt"
	"math/big"
	"os"
	"regexp"
	"strings"
)

func sendFunds() error {
	amount := flag.Arg(1)
	recipient := flag.Arg(2)
	seed, err := getSeedForSending(amount, recipient)
	if err != nil {
		return err
	}
	privateKey := getPrivateKey(seed, uint32(accountIndexFlag))
	address, err := getAddress(privateKey)
	if err != nil {
		return err
	}
	info, err := getAccountInfo(address)
	if err != nil {
		return err
	}
	if info.Frontier == "0000000000000000000000000000000000000000000000000000000000000000" {
		return fmt.Errorf("account has not yet been opened")
	}
	fmt.Fprintf(os.Stderr, "Creating send block... ")
	err = sendFundsToAccount(info, amount, recipient, privateKey)
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "done")
	return nil
}

func getSeedForSending(amount, recipient string) (*big.Int, error) {
	seed, err := getSeed()
	if err != nil {
		return nil, err
	}
	if !yFlag {
		fmt.Printf("Send %s NANO to %s? [y/N]: ", amount, recipient)

		// Explicitly openning /dev/tty ensures function, even if the
		// standard input is not a terminal.
		tty, err := os.Open("/dev/tty")
		if err != nil {
			msg := "could not open terminal for confirmation input: %v"
			return nil, fmt.Errorf(msg, err)
		}
		defer tty.Close()

		var confirmation string
		fmt.Fscanln(tty, &confirmation)
		if confirmation != "y" && confirmation != "Y" {
			fmt.Fprintln(os.Stderr, "Send aborted.")
			os.Exit(0)
		}
	}
	return seed, nil
}

func sendFundsToAccount(info accountInfo, amount, recipient string, privateKey *big.Int) error {
	address, err := getAddress(privateKey)
	if err != nil {
		return err
	}
	balance, err := getBalanceAfterSend(info.Balance, amount)
	if err != nil {
		return err
	}
	recipientNumber, err := getPublicKeyFromAddress(recipient)
	if err != nil {
		return err
	}
	recipientBytes := bigIntToBytes(recipientNumber, 32)
	block := block{
		Type:           "state",
		Account:        address,
		Previous:       info.Frontier,
		Representative: info.Representative,
		Balance:        balance.String(),
		Link:           fmt.Sprintf("%064X", recipientBytes),
	}
	if err = block.sign(privateKey); err != nil {
		return err
	}
	if err = block.addWork(sendWorkThreshold, privateKey); err != nil {
		return err
	}
	process := process{
		Action:    "process",
		JsonBlock: "true",
		Subtype:   "send",
		Block:     block,
	}
	return doProcessRPC(process)
}

func getBalanceAfterSend(oldBalance string, amount string) (*big.Int, error) {
	balance, ok := big.NewInt(0).SetString(oldBalance, 10)
	if !ok {
		err := fmt.Errorf("cannot parse '%s' as an integer", oldBalance)
		return nil, err
	}
	amountRaw, err := nanoStringToRaw(amount)
	if err != nil {
		return nil, err
	}
	return balance.Sub(balance, amountRaw), nil
}

func nanoStringToRaw(amountString string) (*big.Int, error) {
	pattern := `^([0-9]+|[0-9]*\.[0-9]{1,30})$`
	amountOk, err := regexp.MatchString(pattern, amountString)
	if !amountOk {
		return nil, fmt.Errorf("'%s' is no legal amountString", amountString)
	} else if err != nil {
		return nil, err
	}
	missingZerosUntilRaw := 30
	if i := strings.Index(amountString, "."); i > -1 {
		missingZerosUntilRaw -= len(amountString) - i - 1
		amountString = strings.Replace(amountString, ".", "", 1)
	}
	amountString += strings.Repeat("0", missingZerosUntilRaw)
	amount, ok := big.NewInt(0).SetString(amountString, 10)
	if !ok {
		err := fmt.Errorf("cannot parse '%s' as an interger", amountString)
		return nil, err
	}
	return amount, nil
}
