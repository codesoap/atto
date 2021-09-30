package atto

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

// AccountInfo holds the relevant data returned by an account_info RPC
// and the public key and address of the account.
type AccountInfo struct {
	// Ignore this field. It only exists because of
	// https://github.com/nanocurrency/nano-node/issues/1782.
	Error string `json:"error"`

	Frontier       string `json:"frontier"`
	Representative string `json:"representative"`
	Balance        string `json:"balance"`

	PublicKey *big.Int `json:"-"`
	Address   string   `json:"-"`
}

// Send creates a send block, which is hashed but missing the signature
// and work. The Frontier and Balance of the AccountInfo will be
// updated. The amount is interpreted as Nano, not raw!
func (i *AccountInfo) Send(amount, toAddr string) (Block, error) {
	balance, err := getBalanceAfterSend(i.Balance, amount)
	if err != nil {
		return Block{}, err
	}
	recipientNumber, err := getPublicKeyFromAddress(toAddr)
	if err != nil {
		return Block{}, err
	}
	recipientBytes := bigIntToBytes(recipientNumber, 32)
	block := Block{
		Type:           "state",
		SubType:        "send",
		Account:        i.Address,
		PublicKey:      i.PublicKey,
		Previous:       i.Frontier,
		Representative: i.Representative,
		Balance:        balance.String(),
		Link:           fmt.Sprintf("%064X", recipientBytes),
	}
	err = block.hash()
	i.Frontier = block.Hash
	i.Balance = block.Balance
	return block, err
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

// Change creates a change block, which is hashed but missing the
// signature and work. The Frontier of the AccountInfo will be updated.
func (i *AccountInfo) Change(representative string) (Block, error) {
	block := Block{
		Type:           "state",
		SubType:        "change",
		Account:        i.Address,
		PublicKey:      i.PublicKey,
		Previous:       i.Frontier,
		Representative: representative,
		Balance:        i.Balance,
		Link:           "0000000000000000000000000000000000000000000000000000000000000000",
	}
	err := block.hash()
	i.Frontier = block.Hash
	return block, err
}

// Receive creates a receive block, which is hashed but missing the
// signature and work. The Frontier and Balance of the AccountInfo will
// be updated.
func (i *AccountInfo) Receive(pending Pending) (Block, error) {
	updatedBalance, ok := big.NewInt(0).SetString(i.Balance, 10)
	if !ok {
		err := fmt.Errorf("cannot parse '%s' as an integer", i.Balance)
		return Block{}, err
	}
	amount, ok := big.NewInt(0).SetString(pending.Amount, 10)
	if !ok {
		err := fmt.Errorf("cannot parse '%s' as an integer", pending.Amount)
		return Block{}, err
	}
	if amount.Sign() < 1 {
		err := fmt.Errorf("amount '%s' is not positive", pending.Amount)
		return Block{}, err
	}
	updatedBalance.Add(updatedBalance, amount)
	block := Block{
		Type:           "state",
		SubType:        "receive",
		Account:        i.Address,
		PublicKey:      i.PublicKey,
		Previous:       i.Frontier,
		Representative: i.Representative,
		Balance:        updatedBalance.String(),
		Link:           pending.Hash,
	}
	err := block.hash()
	i.Frontier = block.Hash
	i.Balance = block.Balance
	return block, err
}
