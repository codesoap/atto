package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
)

type pending struct {
	Error  string        `json:"error"`
	Blocks pendingBlocks `json:"blocks"`
}

type pendingBlocks map[string]pendingBlockSource

// UnmarshalJSON just unmarshals a list of strings, but
// interprets an empty string as an empty list. This is
// necessary due to a bug in the Nano node implementation. See
// https://github.com/nanocurrency/nano-node/issues/3161.
func (b *pendingBlocks) UnmarshalJSON(in []byte) error {
	if string(in) == `""` {
		return nil
	}
	var raw map[string]pendingBlockSource
	err := json.Unmarshal(in, &raw)
	*b = pendingBlocks(raw)
	return err
}

type pendingBlockSource struct {
	Amount string `json:"amount"`
	Source string `json:"source"`
}

func printBalance() error {
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
	updatedBalance, err := receivePendingSends(info, privateKey)
	if err != nil {
		return err
	}
	fmt.Println(rawToNanoString(updatedBalance))
	return nil
}

func receivePendingSends(info accountInfo, privateKey *big.Int) (updatedBalance *big.Int, err error) {
	updatedBalance, ok := big.NewInt(0).SetString(info.Balance, 10)
	if !ok {
		err = fmt.Errorf("cannot parse '%s' as an integer", info.Balance)
		return
	}
	address, err := getAddress(privateKey)
	if err != nil {
		return
	}
	sends, err := getPendingSends(address)
	if err != nil {
		return
	}
	previousBlock := info.Frontier
	for blockHash, source := range sends {
		amount, ok := big.NewInt(0).SetString(source.Amount, 10)
		if !ok {
			err = fmt.Errorf("cannot parse '%s' as an integer", source.Amount)
			return
		}
		updatedBalance = updatedBalance.Add(updatedBalance, amount)
		txt := "Creating receive block for %s from %s... "
		fmt.Fprintf(os.Stderr, txt, rawToNanoString(amount), source.Source)

		block := block{
			Type:           "state",
			Account:        address,
			Previous:       previousBlock,
			Representative: info.Representative,
			Balance:        updatedBalance.String(),
			Link:           blockHash,
			LinkAsAccount:  source.Source,
		}
		if err = block.sign(privateKey); err != nil {
			return
		}
		if err = block.addWork(receiveWorkThreshold, privateKey); err != nil {
			return
		}
		process := process{
			Action:    "process",
			JsonBlock: "true",
			Subtype:   "receive",
			Block:     block,
		}
		if err = doProcessRPCCall(process); err != nil {
			return
		}

		fmt.Fprintln(os.Stderr, "done")
		previousBlock = block.Hash
	}
	return
}

func getPendingSends(address string) (sends pendingBlocks, err error) {
	requestBody := fmt.Sprintf(`{`+
		`"action": "pending", `+
		`"account": "%s", `+
		`"include_only_confirmed": "true", `+
		`"source": "true"`+
		`}`, address)
	responseBytes, err := doRPC(requestBody)
	if err != nil {
		return
	}
	var pending pending
	err = json.Unmarshal(responseBytes, &pending)
	// Need to check pending.Error because of
	// https://github.com/nanocurrency/nano-node/issues/1782.
	if err == nil && pending.Error != "" {
		err = fmt.Errorf("could not fetch unreceived sends: %s", pending.Error)
	}
	return pending.Blocks, err
}
