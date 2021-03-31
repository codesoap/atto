package main

import (
	"flag"
	"fmt"
	"math/big"
	"os"
)

func changeRepresentative() error {
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
	representative := flag.Arg(1)
	fmt.Fprintf(os.Stderr, "Creating change block... ")
	err = changeRepresatativeOfAccount(info, representative, privateKey)
	if err != nil {
		fmt.Fprintln(os.Stderr, "")
		return err
	}
	fmt.Fprintln(os.Stderr, "done")
	return nil
}

func changeRepresatativeOfAccount(info accountInfo, representative string, privateKey *big.Int) error {
	address, err := getAddress(privateKey)
	if err != nil {
		return err
	}
	block := block{
		Type:           "state",
		Account:        address,
		Previous:       info.Frontier,
		Representative: representative,
		Balance:        info.Balance,
		Link:           "0000000000000000000000000000000000000000000000000000000000000000",
	}
	if err = block.sign(privateKey); err != nil {
		return err
	}
	if err = block.addWork(changeWorkThreshold, privateKey); err != nil {
		return err
	}
	process := process{
		Action:    "process",
		JsonBlock: "true",
		Subtype:   "change",
		Block:     block,
	}
	return doProcessRPCCall(process)
}
