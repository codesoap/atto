package main

import (
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
)

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

func base32Decode(in string) (*big.Int, error) {
	reverseAlphabet := map[rune]*big.Int{}
	reverseAlphabet['1'] = big.NewInt(0)
	reverseAlphabet['3'] = big.NewInt(1)
	reverseAlphabet['4'] = big.NewInt(2)
	reverseAlphabet['5'] = big.NewInt(3)
	reverseAlphabet['6'] = big.NewInt(4)
	reverseAlphabet['7'] = big.NewInt(5)
	reverseAlphabet['8'] = big.NewInt(6)
	reverseAlphabet['9'] = big.NewInt(7)
	reverseAlphabet['a'] = big.NewInt(8)
	reverseAlphabet['b'] = big.NewInt(9)
	reverseAlphabet['c'] = big.NewInt(10)
	reverseAlphabet['d'] = big.NewInt(11)
	reverseAlphabet['e'] = big.NewInt(12)
	reverseAlphabet['f'] = big.NewInt(13)
	reverseAlphabet['g'] = big.NewInt(14)
	reverseAlphabet['h'] = big.NewInt(15)
	reverseAlphabet['i'] = big.NewInt(16)
	reverseAlphabet['j'] = big.NewInt(17)
	reverseAlphabet['k'] = big.NewInt(18)
	reverseAlphabet['m'] = big.NewInt(19)
	reverseAlphabet['n'] = big.NewInt(20)
	reverseAlphabet['o'] = big.NewInt(21)
	reverseAlphabet['p'] = big.NewInt(22)
	reverseAlphabet['q'] = big.NewInt(23)
	reverseAlphabet['r'] = big.NewInt(24)
	reverseAlphabet['s'] = big.NewInt(25)
	reverseAlphabet['t'] = big.NewInt(26)
	reverseAlphabet['u'] = big.NewInt(27)
	reverseAlphabet['w'] = big.NewInt(28)
	reverseAlphabet['x'] = big.NewInt(29)
	reverseAlphabet['y'] = big.NewInt(30)
	reverseAlphabet['z'] = big.NewInt(31)
	out := big.NewInt(0)
	radix := big.NewInt(32)
	for _, r := range in {
		out.Mul(out, radix)
		val, ok := reverseAlphabet[r]
		if !ok {
			return out, fmt.Errorf("'%c' is no legal base32 character", r)
		}
		out.Add(out, val)
	}
	return out, nil
}

func bigIntToBytes(x *big.Int, n int) []byte {
	return x.FillBytes(make([]byte, n, n))
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
