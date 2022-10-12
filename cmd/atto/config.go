package main

var (
	// The node needs to support the work_generate action. See e.g.
	// https://publicnodes.somenano.com to find public nodes or set up your
	// own node. One public alternative would be https://rainstorm.city/api
	node = "https://proxy.nanos.cc/proxy"

	// defaultRepresentative will be set as the representative when
	// opening an accout, but can be changed afterwards. See e.g.
	// https://mynano.ninja/principals to find representatives.
	defaultRepresentative = "nano_1jtx5p8141zjtukz4msp1x93st7nh475f74odj8673qqm96xczmtcnanos1o"
)
