package main

import (
	"flag"
	"fmt"
	"math/big"
	"os"
	"strings"
)

func sendFunds() error {
	seed, err := getSeed()
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
	amount := flag.Arg(1)
	recipient := flag.Arg(2)
	fmt.Fprintf(os.Stderr, "Creating send block... ")
	err = sendFundsToAccount(info, amount, recipient, privateKey)
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "done")
	return nil
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

// getBalanceAfterSend expects oldBalance to be raw and amount to be
// Nano. amount is converted to raw and subtracted from oldBalance, the
// result is returned.
func getBalanceAfterSend(oldBalance string, amount string) (*big.Int, error) {
	balance, ok := big.NewInt(0).SetString(oldBalance, 10)
	if !ok {
		err := fmt.Errorf("cannot parse '%s' as an integer", oldBalance)
		return nil, err
	}
	missingZerosUntilRaw := 30
	i := strings.Index(amount, ".")
	if i > -1 {
		missingZerosUntilRaw -= len(amount) - i - 1
		if missingZerosUntilRaw < 0 {
			return nil, fmt.Errorf("'%s' has too many decimal places", amount)
		}
		amount = strings.Replace(amount, ".", "", 1)
	}
	amount += strings.Repeat("0", missingZerosUntilRaw)
	amountNumber, ok := big.NewInt(0).SetString(amount, 10)
	if !ok {
		err := fmt.Errorf("cannot parse '%s' as an interger", amount)
		return nil, err
	}
	return balance.Sub(balance, amountNumber), nil
}
