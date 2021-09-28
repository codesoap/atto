package atto

import (
	"encoding/json"
	"fmt"
)

type process struct {
	Action    string `json:"action"`
	JsonBlock string `json:"json_block"`
	SubType   string `json:"subtype"`
	Block     Block  `json:"block"`
}

type processResponse struct {
	Error string `json:"error"`
}

func doProcessRPC(process process, node string) error {
	var requestBody, responseBytes []byte
	requestBody, err := json.Marshal(process)
	if err != nil {
		return err
	}
	responseBytes, err = doRPC(string(requestBody), node)
	if err != nil {
		return err
	}
	var processResponse processResponse
	if err = json.Unmarshal(responseBytes, &processResponse); err != nil {
		return err
	}
	// Need to check processResponse.Error because of
	// https://github.com/nanocurrency/nano-node/issues/1782.
	if processResponse.Error != "" {
		err = fmt.Errorf("could not publish block: %s", processResponse.Error)
	}
	return err
}
