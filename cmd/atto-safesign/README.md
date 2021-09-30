# Creating blocks from an offline computer
If you want to keep your seed extra safe you may choose to never take
it onto a computer that is connected to the internet. `atto-safesign`
enables you to do this by creating a file which contains initially
unsigned blocks. The blocks in this file can then be signed on the
offline computer and transferred back to the online computer to submit
the blocks to the Nano network.

To transfer data between the off- and online computer, you can use a
USB thumb drive, but make sure that there is no maleware on the drive.
A safer alternative could be to use QR codes (e.g. with the tools
`qrencode` and `zbarimg`), to transfer the atto file with the blocks to
and from the offline computer after manually checking the contents of
the file.

# Usage
```
online$ # These steps take place on an online computer:
online$ echo $MY_PUBLIC_KEY | atto-safesign test.atto receive
online$ echo $MY_PUBLIC_KEY | atto-safesign test.atto representative nano_3up3y8cd3hhs7zdpmkpssgb1iyjpke3xwmgqy8rg58z1hwryqpjqnkuqayps

offline$ # The sign subcommand can then be used on an offline computer:
offline$ pass nano | atto-safesign test.atto sign

online$ # Back at the online computer, the now signed blocks can be submitted:
online$ atto-safesign test.atto submit
```

```
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
address as the first line of their standard input.

The sign subcommand expects a seed as the first line of standard input.
It also expects manual confirmation before signing blocks, unless the -y
flag is given.

The receive, representative and send subcommands will generate blocks
and append them to FILE. The blocks will still be lacking their
signature.

The sign subcommand will add signatures to all blocks in FILE. It is the
only subcommand that requires no network connection.

The submit subcommand will submit all blocks contained in FILE to the
Nano network.

ACCOUNT_INDEX is an optional parameter, which allows you to use
different accounts derived from the given seed. By default the account
with index 0 is chosen.
```
