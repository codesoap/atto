**WORK IN PROGRESS**

TODO: Some way to check unconfirmed sends?!
TODO: QR-Link for receival?
TODO: QR-Link for sending?

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
  0: nano_3cyb3rwp5ba47t5jdzm5o7apeduppsgzw8ockn1dqt4xcqgapta6gh5htnnh

$ # Choosing a representative is important for keeping the network
$ # decentralized.
$ pass nano | atto representative nano_1jr699mk1fi6mxy1y76fmuyf3dgms8s5pzcsge5cyt1az93x4n18uxjenx93

$ # The balance command will receive pending funds automatically.
$ pass nano | atto balance
1.3370 NANO

$ pass nano | atto send 0.1 nano_11zdqnjpisos53uighoaw95satm4ptdruck7xujbjcs44pbkkbw1h3zomns5
Send 0.1 NANO to nano_11zdqnjpisos53uighoaw95satm4ptdruck7xujbjcs44pbkkbw1h3zomns5? [y/N]: y

$ atto -h
Usage:
	atto n[ew]
	atto [(-n COUNT|-a ACCOUNT_INDEX)] a[ddress]
	atto [--account-index ACCOUNT_INDEX] r[epresentative] REPRESENTATIVE
	atto [--account-index ACCOUNT_INDEX] [--no-receive] b[alance]
	atto [--account-index ACCOUNT_INDEX] [--no-confirm] s[end] AMOUNT RECEIVER
```
