[![GoDoc](https://godoc.org/github.com/codesoap/atto?status.svg)](https://godoc.org/github.com/codesoap/atto)

atto is a tiny Nano wallet, which focuses on ease of use through
simplicity. Included is a rudimentary Go library to interact with Nano
nodes.

Disclaimer: I am no cryptographer and atto has not been audited. I
cannot guarantee that atto is free of security compromising bugs. If
you want to be extra cautious, I recommend offline signing, which is
possible with the included [atto-safesign](cmd/atto-safesign/).

# Installation
You can download precompiled binaries from the [releases
page](https://github.com/codesoap/atto/releases) or build atto yourself
like this; go 1.15 or higher is required:

```shell
git clone 'https://github.com/codesoap/atto.git'
cd atto
go build ./cmd/atto
# The atto binary is now available at ./atto. You could also install
# to ~/go/bin/ by executing "go install ./cmd/atto".
```

For Arch Linux @kseistrup also made [atto available in the
AUR](https://aur.archlinux.org/packages/atto/).

# Usage
```console
$ # The new command generates a new seed.
$ atto new
D420296F5FEF486175FAA8F649DED00A5B0A096DB8D03972937542C51A7F296C
$ # Store it in your password manager:
$ pass insert nano
Enter password for nano: D420296F5FEF486175FAA8F649DED00A5B0A096DB8D03972937542C51A7F296C
Retype password for nano: D420296F5FEF486175FAA8F649DED00A5B0A096DB8D03972937542C51A7F296C

$ # The address command shows the address for an account.
$ pass nano | atto address
nano_3cyb3rwp5ba47t5jdzm5o7apeduppsgzw8ockn1dqt4xcqgapta6gh5htnnh

$ # With address and all following commands you can also provide an
$ # alternative account index (default is 0):
$ pass nano | atto -a 1 address
nano_1o3igdpf8c4msdgwcop71x4o16zzkhe4kyku4axdi8iwh8wh13e4fwgherik

$ # The balance command will receive receivable funds automatically.
$ pass nano | atto balance
Creating receive block for 1.025 from nano_34ymtnmhwseiex4eqf7nnf5wcyg44kknuuen5wwurm18ma91msf6e1pqo8hx... done
Creating receive block for 0.1 from nano_39nd8eksw1ia6aokn96z4uthocke47hfsx9gr31othm1nrfwnzmmaeehiccq... done
1.337 NANO

$ # Choosing a representative is important for keeping the network
$ # decentralized.
$ pass nano | atto representative nano_1jr699mk1fi6mxy1y76fmuyf3dgms8s5pzcsge5cyt1az93x4n18uxjenx93
Creating change block... done

$ # To avoid accidental loss of funds, the send command requires
$ # confirmation, unless the -y flag is given:
$ pass nano | atto send 0.1 nano_11zdqnjpisos53uighoaw95satm4ptdruck7xujbjcs44pbkkbw1h3zomns5
Send 0.1 NANO to nano_11zdqnjpisos53uighoaw95satm4ptdruck7xujbjcs44pbkkbw1h3zomns5? [y/N]: y
Creating send block... done

$ atto -h
Usage:
	atto -v
	atto n[ew]
	atto [-a ACCOUNT_INDEX] a[ddress]
	atto [-a ACCOUNT_INDEX] b[alance]
	atto [-a ACCOUNT_INDEX] r[epresentative] [NEW_REPRESENTATIVE]
	atto [-a ACCOUNT_INDEX] [-y] s[end] AMOUNT RECEIVER

If the -v flag is provided, atto will print its version number.

The new subcommand generates a new seed, which can later be used with
the other subcommands.

The address, balance, representative and send subcommands expect a seed
as the first line of their standard input. Showing the first address of
a newly generated key could work like this:
atto new | tee seed.txt | atto address

The send subcommand also expects manual confirmation of the transaction,
unless the -y flag is given.

The address subcommand displays addresses for a seed, the balance
subcommand receives receivable blocks and shows the balance of an
account, the representative subcommand shows the current representative
if NEW_REPRESENTATIVE is not given and changes the account's
representative if it is given and the send subcommand sends funds to an
address.

ACCOUNT_INDEX is an optional parameter, which must be a number between 0
and 4,294,967,295. It allows you to use multiple accounts derived from
the same seed. By default the account with index 0 is chosen.

Environment:
	ATTO_BASIC_AUTH_USERNAME  The username for HTTP Basic Authentication.
	                          If set, HTTP Basic Authentication will be
	                          used when making requests to the node.
	ATTO_BASIC_AUTH_PASSWORD  The password to use for HTTP Basic
	                          Authentication.
```

# Technical details
atto is written with ca. 1000 lines of code and uses minimal external
dependencies. This makes it easy to audit the code yourself and ensure,
that it does nothing you wouldn't want it to do.

To change some defaults, like the node to use, take a look at
`cmd/atto/config.go`.

Signatures are created without the help of a node, to avoid your seed or
private keys being stolen by a node operator. The received account info
is always validated using block signatures to ensure the node operator
cannot manipulate atto by, for example, reporting wrong balances.

atto does not have any persistance and writes nothing to your
file system. This makes atto very portable, but also means, that
no history is stored locally. I recommend using a service like
https://nanocrawler.cc/ to investigate transaction history.

# Donations
If you want to show your appreciation for atto, you can donate to me at
`nano_1i7wsbehgwhxct91wpojr1j588ydikd64uc7p3kj54nofqioc6ydjopezf13`.
