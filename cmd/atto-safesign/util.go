package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/big"
	"os"
	"runtime"
	"strings"

	"github.com/codesoap/atto"
)

func getFirstStdinLine() (string, error) {
	in := bufio.NewReader(os.Stdin)
	firstLine, err := in.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(firstLine), nil
}

// getLatestAccountInfo returns an atto.AccountInfo with the latest
// available block as it's Frontier. This is either the last block from
// the file or the one fetched from the network, if the file contains no
// blocks.
func getLatestAccountInfo(acc atto.Account) (atto.AccountInfo, error) {
	blocks, err := getBlocksFromFile()
	if err != nil {
		return atto.AccountInfo{}, err
	}
	if len(blocks) == 0 {
		return acc.FetchAccountInfo(node)
	}
	latestBlock := blocks[len(blocks)-1]
	hash, err := latestBlock.Hash()
	if err != nil {
		return atto.AccountInfo{}, err
	}
	info := atto.AccountInfo{
		Frontier:       hash,
		Representative: latestBlock.Representative,
		Balance:        latestBlock.Balance,
		PublicKey:      acc.PublicKey,
		Address:        acc.Address,
	}
	return info, nil
}

func getBlocksFromFile() ([]atto.Block, error) {
	file, err := os.Open(flag.Arg(0))
	if err != nil {
		// The file has not been found, which is OK.
		return []atto.Block{}, nil
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	blocks := make([]atto.Block, 0)
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		var block atto.Block
		if err = json.Unmarshal(line, &block); err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}

func appendLineToFile(in []byte) error {
	file, err := os.OpenFile(flag.Arg(0), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err = file.Write(in); err != nil {
		return err
	}
	_, err = file.Write([]byte{'\n'})
	return err
}

func rawToNanoString(raw *big.Int) string {
	rawPerNano, _ := big.NewInt(0).SetString("100000000000000000000000000000", 10)
	absRaw := big.NewInt(0).Abs(raw)
	integerDigits, fractionalDigits := big.NewInt(0).QuoRem(absRaw, rawPerNano, big.NewInt(0))
	res := integerDigits.String()
	if fractionalDigits.Sign() != 0 {
		fractionalDigitsString := fmt.Sprintf("%029s", fractionalDigits.String())
		res += "." + strings.TrimRight(fractionalDigitsString, "0")
	}
	if raw.Sign() < 0 {
		return "-" + res + " NANO"
	}
	return res + " NANO"
}

func letUserVerifyBlock(block atto.Block) (err error) {
	if !yFlag {
		balanceInt, ok := big.NewInt(0).SetString(block.Balance, 10)
		if !ok {
			return fmt.Errorf("cannot parse '%s' as an integer", block.Balance)
		}
		balanceNano := rawToNanoString(balanceInt)
		txt := "Sign block that sets balance to %s and representative to %s? [y/N]: "
		fmt.Fprintf(os.Stderr, txt, balanceNano, block.Representative)

		// Explicitly openning /dev/tty or CONIN$ ensures function, even if
		// the standard input is not a terminal.
		var tty *os.File
		if runtime.GOOS == "windows" {
			tty, err = os.Open("CONIN$")
		} else {
			tty, err = os.Open("/dev/tty")
		}
		if err != nil {
			msg := "could not open terminal for confirmation input: %v"
			return fmt.Errorf(msg, err)
		}
		defer tty.Close()

		var confirmation string
		fmt.Fscanln(tty, &confirmation)
		if confirmation != "y" && confirmation != "Y" {
			fmt.Fprintln(os.Stderr, "Signing aborted.")
			os.Exit(0)
		}
	}
	return
}
