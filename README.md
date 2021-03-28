**WORK IN PROGRESS**

atto is a tiny Nano wallet, targeted at friends of the UNIX philosophy.
The design goal is to provide a minimal set of actions to manipulate
your Nano account. atto provides no means of inspecting transaction
history as it is completely stateless. Use a Nano crawler like
https://nanocrawler.cc/ if you need such info.

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
```
