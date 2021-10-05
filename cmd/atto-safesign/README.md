`atto-safesign` is intended to be used as an extension to `atto`, so I
strongly recommend you familiarize yourself with `atto` before looking
at `atto-safesign`.

# Motivation
If you want to keep your seed extra safe you may choose to never take
it onto a computer that is connected to the internet. `atto-safesign`
enables you to do this by creating a file which contains initially
unsigned blocks. The blocks in this file can then be signed on the
offline computer and transferred back to the online computer to submit
the blocks to the Nano network.

# Installation
You can download precompiled binaries from the [releases
page](https://github.com/codesoap/atto/releases) or build atto-safesign
yourself like this; go 1.15 or higher is required:

```shell
git clone 'https://github.com/codesoap/atto.git'
cd atto
go build ./cmd/atto-safesign/
# The atto-safesign binary is now available at ./atto-safesign. You could
# also install to ~/go/bin/ by executing "go install ./cmd/atto-safesign/".
```

# Usage
Here is an example use case where pending sends are received and the
representative changed:

```
online$ # These steps take place on an online computer:
online$ echo $MY_PUBLIC_KEY | atto-safesign test.atto receive
online$ echo $MY_PUBLIC_KEY | atto-safesign test.atto representative nano_3up3y8cd3hhs7zdpmkpssgb1iyjpke3xwmgqy8rg58z1hwryqpjqnkuqayps

offline$ # The sign subcommand can then be used on an offline computer:
offline$ pass nano | atto-safesign test.atto sign

online$ # Back at the online computer, the now signed blocks can be submitted:
online$ echo $MY_PUBLIC_KEY | atto-safesign test.atto submit
```

This is `atto-safesign`'s help text:
```console
$ atto-safesign -h
Usage:
        atto-safesign -v
        atto-safesign FILE receive
        atto-safesign FILE representative REPRESENTATIVE
        atto-safesign FILE send AMOUNT RECEIVER
        atto-safesign [-a ACCOUNT_INDEX] [-y] FILE sign
        atto-safesign FILE submit

If the -v flag is provided, atto-safesign will print its version number.

The receive, representative, send and submit subcommands expect a Nano
address as the first line of their standard input. This address will be
the account of the generated and submitted blocks.

The receive, representative and send subcommands will generate blocks
and append them to FILE. The blocks will still be lacking their
signature. The receive subcommand will create multiple blocks, if there
are multiple pending sends that can be received. The representative
subcommand will create a block for changing the representative and the
send subcommand will create a block for sending funds to an address.

The sign subcommand expects a seed as the first line of standard input.
It also expects manual confirmation before signing blocks, unless the
-y flag is given. The seed and ACCOUNT_INDEX must belong to the address
used when creating blocks with receive, representative or send.

The sign subcommand will add signatures to all blocks in FILE. It is the
only subcommand that requires no network connection.

The submit subcommand will submit all blocks contained in FILE to the
Nano network.

ACCOUNT_INDEX is an optional parameter, which allows you to use
different accounts derived from the given seed. By default the account
with index 0 is chosen.
```
