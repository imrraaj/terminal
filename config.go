package main

import (
	"crypto/ecdsa"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonirico/go-hyperliquid"
)

type Config struct {
	URL        string
	PrivateKey *ecdsa.PrivateKey
	Address    string
	RedisURL   string
}

func NewConfig() Config {
	pk, err := os.ReadFile(".secret")
	if err != nil {
		panic(fmt.Errorf("failed to read .secret file: %w", err))
	}
	privateKey, err := crypto.HexToECDSA(string(pk))
	if err != nil {
		panic(fmt.Errorf("invalid private key: %w", err))
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		panic(fmt.Errorf("failed to cast public key to ECDSA"))
	}
	return Config{
		URL:        hyperliquid.TestnetAPIURL,
		PrivateKey: privateKey,
		Address:    crypto.PubkeyToAddress(*publicKeyECDSA).Hex(),
		RedisURL:   "localhost:6379",
	}
}

func (c *Config) SetSourceURL(url string) {
	c.URL = url
}
