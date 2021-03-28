package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"os"
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
	publicKey, err := getPublicKeyFromAddress(address)
	if err != nil {
		return err
	}

	balance, ok := big.NewInt(0).SetString(info.Balance, 10)
	if !ok {
		err = fmt.Errorf("cannot parse '%s' as an integer", info.Balance)
		return err
	}
	amountNumber, ok := big.NewFloat(0).SetString(amount)
	if !ok {
		err = fmt.Errorf("cannot parse '%s' as a number", amount)
		return err
	}
	rawPerNano, _ := big.NewFloat(0).SetString("1000000000000000000000000000000")
	amountNumber = amountNumber.Mul(amountNumber, rawPerNano)
	if !amountNumber.IsInt() {
		err = fmt.Errorf("'%s' is no legal amount", amount)
		return err
	}
	amountInt, _ := amountNumber.Int(big.NewInt(0))
	balance = balance.Sub(balance, amountInt)

	publicKeyBytes := make([]byte, 32, 32)
	publicKey.FillBytes(publicKeyBytes)
	block := block{
		Type:           "state",
		Account:        address,
		Previous:       info.Frontier,
		Representative: info.Representative,
		Balance:        balance.String(),
		Link:           fmt.Sprintf("%064X", publicKeyBytes),
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
	var requestBody, responseBytes []byte
	requestBody, err = json.Marshal(process)
	if err != nil {
		return err
	}
	responseBytes, err = doRPC(string(requestBody))
	if err != nil {
		return err
	}
	var processResponse processResponse
	err = json.Unmarshal(responseBytes, &processResponse)
	if err != nil {
		return err
	}
	// Need to check pending.Error because of
	// https://github.com/nanocurrency/nano-node/issues/1782.
	if processResponse.Error != "" {
		err = fmt.Errorf("could not publish send block: %s", processResponse.Error)
		return err
	}
	return err
}
