package atto

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"

	"golang.org/x/crypto/blake2b"
)

// GenerateSeed generates a new random seed.
func GenerateSeed() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	return fmt.Sprintf("%X", b), err
}

// NewPrivateKey creates a private key from the given seed and index.
func NewPrivateKey(seed string, index uint32) (*big.Int, error) {
	seedInt, ok := big.NewInt(0).SetString(seed, 16)
	if !ok {
		return nil, fmt.Errorf("could not parse seed")
	}
	seedBytes := bigIntToBytes(seedInt, 32)
	indexBytes := bigIntToBytes(big.NewInt(int64(index)), 4)
	in := append(seedBytes, indexBytes...)
	privateKeyBytes := blake2b.Sum256(in)
	return big.NewInt(0).SetBytes(privateKeyBytes[:]), nil
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

func doRPC(requestBody, node string) (responseBytes []byte, err error) {
	req, err := http.NewRequest("POST", node, strings.NewReader(requestBody))
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")
	if RequestInterceptor != nil {
		if err = RequestInterceptor(req); err != nil {
			err = fmt.Errorf("request interceptor failed: %v", err)
			return
		}
	}
	resp, err := http.DefaultClient.Do(req)
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

func getPublicKeyFromAddress(address string) (*big.Int, error) {
	if len(address) == 64 {
		return base32Decode(address[4:56])
	} else if len(address) == 65 {
		return base32Decode(address[5:57])
	}
	return nil, fmt.Errorf("could not parse address %s", address)
}
