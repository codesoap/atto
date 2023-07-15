package atto

import (
	"encoding/json"
)

// Receivable represents a block that is waiting to be received.
type Receivable struct {
	Hash   string
	Amount string
	Source string
}

type internalReceivable struct {
	Error  string           `json:"error"`
	Blocks receivableBlocks `json:"blocks"`
}

type receivableBlocks map[string]receivableBlock

// UnmarshalJSON just unmarshals a list of strings, but
// interprets an empty string as an empty list. This is
// necessary due to a bug in the Nano node implementation. See
// https://github.com/nanocurrency/nano-node/issues/3161.
func (b *receivableBlocks) UnmarshalJSON(in []byte) error {
	if string(in) == `""` {
		return nil
	}
	var raw map[string]receivableBlock
	err := json.Unmarshal(in, &raw)
	*b = receivableBlocks(raw)
	return err
}

type receivableBlock struct {
	Amount string `json:"amount"`
	Source string `json:"source"`
}

func internalReceivableToReceivable(internalReceivable internalReceivable) []Receivable {
	receivables := make([]Receivable, 0)
	for hash, source := range internalReceivable.Blocks {
		receivable := Receivable{hash, source.Amount, source.Source}
		receivables = append(receivables, receivable)
	}
	return receivables
}
