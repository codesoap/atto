package main

var (
	// The node needs to support the work_generate action:
	nodeUrl = "https://proxy.powernode.cc/proxy"

	defaultRepresentative = "nano_18shbirtzhmkf7166h39nowj9c9zrpufeg75bkbyoobqwf1iu3srfm9eo3pz"

	sendWorkThreshold    uint64 = 0xfffffff800000000
	changeWorkThreshold  uint64 = 0xfffffff800000000
	receiveWorkThreshold uint64 = 0xfffffe0000000000
)
