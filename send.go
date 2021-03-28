package main

import (
	"flag"
	"fmt"
	"math/big"
	"os"
)

// TODO: Ask for confirmation

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
	fmt.Fprintf(os.Stderr, "Creating send block (may take many minutes)... ")
	err = sendFundsToAccount(info, amount, recipient, privateKey)
	if err != nil {
		fmt.Fprintln(os.Stderr, "")
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
	recipientBytes := make([]byte, 32, 32)
	recipientNumber, err := getPublicKeyFromAddress(recipient)
	if err != nil {
		return err
	}
	recipientNumber.FillBytes(recipientBytes)
	block := block{
		Type:           "state",
		Account:        address,
		Previous:       info.Frontier,
		Representative: info.Representative,
		Balance:        balance.String(),
		Link:           fmt.Sprintf("%064X", recipientBytes),
		LinkAsAccount:  recipient,
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
	return doProcessRPCCall(process)
}

func getBalanceAfterSend(oldBalance string, amount string) (*big.Int, error) {
	balance, ok := big.NewInt(0).SetString(oldBalance, 10)
	if !ok {
		err := fmt.Errorf("cannot parse '%s' as an integer", oldBalance)
		return nil, err
	}
	amountNumber, ok := big.NewFloat(0).SetString(amount)
	if !ok {
		err := fmt.Errorf("cannot parse '%s' as a number", amount)
		return nil, err
	}
	rawPerNano, _ := big.NewFloat(0).SetString("1000000000000000000000000000000")
	amountNumber = amountNumber.Mul(amountNumber, rawPerNano)
	if !amountNumber.IsInt() {
		err := fmt.Errorf("'%s' is no legal amount", amount)
		return nil, err
	}
	amountInt, _ := amountNumber.Int(big.NewInt(0))
	return balance.Sub(balance, amountInt), nil
}
