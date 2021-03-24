package main

type process struct {
	Action    string `json:"action"`
	JsonBlock string `json:"json_block"`
	Subtype   string `json:"subtype"`
	Block     block  `json:"block"`
}

type block struct {
	Type           string `json:"type"`
	Account        string `json:"account"`
	Previous       string `json:"previous"`
	Representative string `json:"representative"`
	Balance        string `json:"balance"`
	Link           string `json:"link"`
	LinkAsAccount  string `json:"link_as_account"`
	Signature      string `json:"signature"`
	Work           string `json:"work"`
}
