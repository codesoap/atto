package atto

import (
	"encoding/json"
)

// Pending represents a block that is waiting to be received.
type Pending struct {
	Hash   string
	Amount string
	Source string
}

type internalPending struct {
	Error  string        `json:"error"`
	Blocks pendingBlocks `json:"blocks"`
}

type pendingBlocks map[string]pendingBlock

// UnmarshalJSON just unmarshals a list of strings, but
// interprets an empty string as an empty list. This is
// necessary due to a bug in the Nano node implementation. See
// https://github.com/nanocurrency/nano-node/issues/3161.
func (b *pendingBlocks) UnmarshalJSON(in []byte) error {
	if string(in) == `""` {
		return nil
	}
	var raw map[string]pendingBlock
	err := json.Unmarshal(in, &raw)
	*b = pendingBlocks(raw)
	return err
}

type pendingBlock struct {
	Amount string `json:"amount"`
	Source string `json:"source"`
}

func internalPendingToPending(internalPending internalPending) []Pending {
	pendings := make([]Pending, 0)
	for hash, source := range internalPending.Blocks {
		pending := Pending{hash, source.Amount, source.Source}
		pendings = append(pendings, pending)
	}
	return pendings
}
