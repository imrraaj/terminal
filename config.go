package main

import (
	"fmt"
	"os"

	"github.com/sonirico/go-hyperliquid"
)

type Config struct {
	URL        string //wether testnet or mainnet
	PrivateKey string // private key to trade with
}

func NewConfig() *Config {
	pk, err := os.ReadFile(".secret")
	if err != nil {
		panic(fmt.Errorf("No private key file found. Please put the privat key in .secret file"))
	}
	return &Config{
		URL:        hyperliquid.TestnetAPIURL,
		PrivateKey: string(pk),
	}
}

func (c *Config) SetSourceURL(url string) {
	c.URL = url
}

func (c *Config) SetPrivateKey(pk string) {
	c.PrivateKey = pk
}
