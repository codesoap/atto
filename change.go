package main

import (
	"flag"
	"fmt"
	"os"
)

func changeRepresentative() error {
	account, err := ownAccount()
	if err != nil {
		return err
	}
	info, err := account.getInfo()
	if err != nil {
		return err
	}
	representative := flag.Arg(1)
	fmt.Fprintf(os.Stderr, "Creating change block... ")
	err = account.changeRepresentativeOfAccount(info, representative)
	if err != nil {
		fmt.Fprintln(os.Stderr, "")
		return err
	}
	fmt.Fprintln(os.Stderr, "done")
	return nil
}

func (a account) changeRepresentativeOfAccount(info accountInfo, representative string) error {
	block := block{
		Type:           "state",
		Account:        a.address,
		Previous:       info.Frontier,
		Representative: representative,
		Balance:        info.Balance,
		Link:           "0000000000000000000000000000000000000000000000000000000000000000",
	}
	if err := block.sign(a); err != nil {
		return err
	}
	if err := block.addWork(changeWorkThreshold, a); err != nil {
		return err
	}
	process := process{
		Action:    "process",
		JsonBlock: "true",
		Subtype:   "change",
		Block:     block,
	}
	return doProcessRPC(process)
}
