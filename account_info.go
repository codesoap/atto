package atto

import (
	"fmt"
	"math/big"
	"strings"
)

// AccountInfo holds the basic data needed for Block creation.
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

// Send creates a send block, which will still be missing its signature
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
		SubType:        SubTypeSend,
		Account:        i.Address,
		Previous:       i.Frontier,
		Representative: i.Representative,
		Balance:        balance.String(),
		Link:           fmt.Sprintf("%064X", recipientBytes),
	}
	hash, err := block.Hash()
	if err != nil {
		return Block{}, err
	}
	i.Frontier = hash
	i.Balance = block.Balance
	return block, err
}

func getBalanceAfterSend(oldBalance string, amount string) (*big.Int, error) {
	balance, ok := big.NewInt(0).SetString(oldBalance, 10)
	if !ok {
		err := fmt.Errorf("cannot parse '%s' as an integer", oldBalance)
		return nil, err
	}
	amountRaw, err := nanoToRaw(amount)
	if err != nil {
		return nil, err
	}
	return balance.Sub(balance, amountRaw), nil
}

func nanoToRaw(amountString string) (*big.Int, error) {
	i := strings.Index(amountString, ".")
	missingZerosUntilRaw := 30
	if i > -1 {
		missingZerosUntilRaw = 31 + i - len(amountString)
		amountString = amountString[:i] + amountString[i+1:] // Remove "."
	}
	amountString += strings.Repeat("0", missingZerosUntilRaw)
	amount, ok := big.NewInt(0).SetString(amountString, 10)
	if !ok {
		return nil, fmt.Errorf("cannot parse '%s' as an interger", amountString)
	}
	return amount, nil
}

// Change creates a change block, which will still be missing its
// signature and work. The Frontier and Representative of the
// AccountInfo will be updated.
func (i *AccountInfo) Change(representative string) (Block, error) {
	block := Block{
		Type:           "state",
		SubType:        SubTypeChange,
		Account:        i.Address,
		Previous:       i.Frontier,
		Representative: representative,
		Balance:        i.Balance,
		Link:           "0000000000000000000000000000000000000000000000000000000000000000",
	}
	hash, err := block.Hash()
	if err != nil {
		return Block{}, err
	}
	i.Frontier = hash
	return block, err
}

// Receive creates a receive block, which will still be missing its
// signature and work. The Frontier and Balance of the AccountInfo will
// be updated.
func (i *AccountInfo) Receive(receivable Receivable) (Block, error) {
	updatedBalance, ok := big.NewInt(0).SetString(i.Balance, 10)
	if !ok {
		err := fmt.Errorf("cannot parse '%s' as an integer", i.Balance)
		return Block{}, err
	}
	amount, ok := big.NewInt(0).SetString(receivable.Amount, 10)
	if !ok {
		err := fmt.Errorf("cannot parse '%s' as an integer", receivable.Amount)
		return Block{}, err
	}
	if amount.Sign() < 1 {
		err := fmt.Errorf("amount '%s' is not positive", receivable.Amount)
		return Block{}, err
	}
	updatedBalance.Add(updatedBalance, amount)
	block := Block{
		Type:           "state",
		SubType:        SubTypeReceive,
		Account:        i.Address,
		Previous:       i.Frontier,
		Representative: i.Representative,
		Balance:        updatedBalance.String(),
		Link:           receivable.Hash,
	}
	hash, err := block.Hash()
	if err != nil {
		return Block{}, err
	}
	i.Frontier = hash
	i.Balance = block.Balance
	return block, err
}
