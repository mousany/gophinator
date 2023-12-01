package runtime

import (
	"crypto/rand"
	"fmt"
)

func newHostname() (string, error) {
	var names = &[]string{
		"cat",
		"world",
		"coffee",
		"girl",
		"man",
		"book",
		"pinguin",
		"moon",
	}
	var adjs = &[]string{
		"red",
		"blue",
		"green",
		"yellow",
		"big",
		"small",
		"tall",
		"thin",
		"round",
		"square",
		"triangular",
		"weird",
		"noisy",
		"silent",
		"soft",
		"irregular",
	}

	idx, err := randomInt()
	if err != nil {
		return "", err
	}
	name := (*names)[idx%len(*names)]

	idx, err = randomInt()
	if err != nil {
		return "", err
	}
	adj := (*adjs)[idx%len(*adjs)]

	idx, err = randomInt()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%s-%d", adj, name, idx), nil
}

func randomInt() (int, error) {
	var buf [4]byte
	_, err := rand.Read(buf[:])
	if err != nil {
		return 0, err
	}

	return int(buf[0])<<24 | int(buf[1])<<16 | int(buf[2])<<8 | int(buf[3]), nil
}
