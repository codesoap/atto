atto is a tiny Nano wallet, which focuses on ease of use through
simplicity. 

Disclaimer: I am no cryptographer and atto has not been audited. I
cannot guarantee that atto is free of security compromising bugs.

# Installation
I recommend building atto yourself, but you can also
download precompiled binaries from the [releases
page](https://github.com/codesoap/atto/releases). Building atto yourself
works like this; go 1.15 or higher is required:

```shell
git clone 'https://github.com/codesoap/atto.git'
cd atto
go build
# The atto binary is now available at ./atto.
# You could also install to ~/go/bin/ by executing "go install".
```

# Usage
```console
$ # The new command generates a new seed.
$ atto new
D420296F5FEF486175FAA8F649DED00A5B0A096DB8D03972937542C51A7F296C
$ # Store it in your password manager:
$ pass insert nano
Enter password for nano: D420296F5FEF486175FAA8F649DED00A5B0A096DB8D03972937542C51A7F296C
Retype password for nano: D420296F5FEF486175FAA8F649DED00A5B0A096DB8D03972937542C51A7F296C

$ # The address command shows the address for a seed and account index.
$ pass nano | atto address
nano_3cyb3rwp5ba47t5jdzm5o7apeduppsgzw8ockn1dqt4xcqgapta6gh5htnnh

$ # The balance command will receive pending funds automatically.
$ pass nano | atto balance
Creating receive block for 1.025 from nano_34ymtnmhwseiex4eqf7nnf5wcyg44kknuuen5wwurm18ma91msf6e1pqo8hx... done
Creating receive block for 0.100 from nano_39nd8eksw1ia6aokn96z4uthocke47hfsx9gr31othm1nrfwnzmmaeehiccq... done
1.337 NANO

$ # Choosing a representative is important for keeping the network
$ # decentralized.
$ pass nano | atto representative nano_1jr699mk1fi6mxy1y76fmuyf3dgms8s5pzcsge5cyt1az93x4n18uxjenx93
Creating change block (may take many minutes)... done

$ # Careful with the send subcommand: No confirmation is required!
$ pass nano | atto send 0.1 nano_11zdqnjpisos53uighoaw95satm4ptdruck7xujbjcs44pbkkbw1h3zomns5
Creating send block (may take many minutes)... done

$ atto -h
Usage:
	atto n[ew]
	atto [-a ACCOUNT_INDEX] a[ddress]
	atto [-a ACCOUNT_INDEX] b[alance]
	atto [-a ACCOUNT_INDEX] r[epresentative] REPRESENTATIVE
	atto [-a ACCOUNT_INDEX] s[end] AMOUNT RECEIVER

The new subcommand generates a new seed, which can later be used with
the other subcommands.

The address, balance, representative and send subcommands will expect
a seed as as the first line of their standard input. Showing the first
address of a newly generated key could work like this:
atto new | tee seed.txt | atto address

The address subcommand displays addresses for a seed, the balance
subcommand receives pending sends and shows the balance of an account,
the representative subcommand changes the account's representative and
the send subcommand sends funds to an address.

ACCOUNT_INDEX is an optional parameter, which must be a number between 0
and 4,294,967,295. It allows you to use multiple accounts derived from
the same seed. By default the account with index 0 is chosen.
```

# Technical details
atto is written with less than 1000 lines of code and uses minimal
external dependencies. This makes it easy to audit the code yourself
and ensure, that it does nothing you wouldn't want it to do.

To change some defaults, take a look at `config.go`.

Signatures and proof of work are created without the help of external
services. This means that the creation of new blocks will take some
time and CPU power. Depending on your CPU, operating system and chance,
calculating the proof of work for "send" and "change" blocks can take
anywhere from a few seconds to 15 minutes or more.

atto does not have any persistance and writes nothing to your
file system. This makes atto very portable, but also means, that
no history is stored locally. I recommend using a service like
https://nanocrawler.cc/ to investigate transaction history.
