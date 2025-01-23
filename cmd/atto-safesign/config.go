package main

var (
	// The node needs to support the work_generate action. See
	// e.g. https://publicnodes.somenano.com to find public nodes
	// or set up your own node.
	node = "https://rainstorm.city/api"

	// defaultRepresentative will be set as the representative when
	// opening an accout, but can be changed afterwards. See e.g.
	// https://blocklattice.io/representatives to find representatives.
	defaultRepresentative = "nano_1jtx5p8141zjtukz4msp1x93st7nh475f74odj8673qqm96xczmtcnanos1o"

	// workSource specifies where the work for block submission shall
	// come from. These options are available:
	// - workSourceLocal: The work is generated on the CPU of the
	//   current computer.
	// - workSourceNode: The work is fetched from the node using the
	//   work_generate action. Make sure that your node supports it.
	// - workSourceLocalFallback: It is attempted to fetch the work
	//   from the node, but if this fails, it will be generated on
	//   the CPU of the current computer.
	workSource = workSourceLocalFallback
)
