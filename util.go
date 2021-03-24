package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"strings"
)

// getSeed takes the first line of the standard input and interprets it
// as a hexadecimal representation of a 32byte seed.
func getSeed() (*big.Int, error) {
	in := bufio.NewReader(os.Stdin)
	firstLine, err := in.ReadString('\n')
	if err != nil {
		return big.NewInt(0), err
	}
	seed, ok := big.NewInt(0).SetString(strings.TrimSpace(firstLine), 16)
	if !ok {
		return seed, errors.New("could not parse seed")
	}
	return seed, nil
}

func base32Encode(in *big.Int) string {
	alphabet := []byte("13456789abcdefghijkmnopqrstuwxyz")
	bigZero := big.NewInt(0)
	bigRadix := big.NewInt(32)
	num := big.NewInt(0).SetBytes(in.Bytes())
	out := make([]byte, 0)
	mod := new(big.Int)
	for num.Cmp(bigZero) > 0 {
		num.DivMod(num, bigRadix, mod)
		out = append(out, alphabet[mod.Int64()])
	}
	for i := 0; i < len(out)/2; i++ {
		out[i], out[len(out)-1-i] = out[len(out)-1-i], out[i]
	}
	return string(out)
}

func revertBytes(in []byte) []byte {
	for i := 0; i < len(in)/2; i++ {
		in[i], in[len(in)-1-i] = in[len(in)-1-i], in[i]
	}
	return in
}

func doRPC(requestBody string) (responseBytes []byte, err error) {
	resp, err := http.Post(nodeUrl, "application/json", strings.NewReader(requestBody))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("received unexpected HTTP return code %d", resp.StatusCode)
		return
	}
	return ioutil.ReadAll(resp.Body)
}

func rawToNanoString(raw *big.Int) string {
	rawPerKnano, _ := big.NewInt(0).SetString("1000000000000000000000000000", 10)
	balance := big.NewInt(0).Div(raw, rawPerKnano).Uint64()
	return fmt.Sprintf("%d.%03d NANO", balance/1000, balance%1000)
}
