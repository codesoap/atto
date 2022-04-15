package main

var (
	// The node needs to support the work_generate action.
	// See e.g. https://publicnodes.somenano.com to find public nodes or
	// set up your own node to use. Here are some public alternatives:
	// - https://proxy.nanos.cc/proxy
	// - https://mynano.ninja/api/node
	// - https://rainstorm.city/api
	node = "https://proxy.powernode.cc/proxy"

	// defaultRepresentative will be set as the representative when
	// opening an accout, but can be changed afterwards. See e.g.
	// https://mynano.ninja/principals to find representatives.
	defaultRepresentative = "ban_3hjo1cehsxrssawmpew98u4ug8bxy4ppht5ch647zpuscdgedfy1xh4yga7z"
)
