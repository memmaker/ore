package main

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"strconv"
)

func DigestHash(s string) []byte {
	prefix := []byte("\x19Ethereum Signed Message:\n")
	messageLength := strconv.Itoa(len(s))
	length := []byte(messageLength)
	sourceString := append(prefix, length...)
	sourceString = append(sourceString, s...)
	h := crypto.Keccak256Hash(sourceString)
	return h.Bytes()
}

type Account struct {
	Name       string
	privateKey *ecdsa.PrivateKey
}

func (a *Account) PublicKey() string {
	publicKeyBytes := crypto.FromECDSAPub(&a.privateKey.PublicKey)
	publicKeyString := hexutil.Encode(publicKeyBytes)
	return publicKeyString
}

func (a *Account) PrivateKey() string {
	privateKeyBytes := crypto.FromECDSA(a.privateKey)
	privateKeyString := hexutil.Encode(privateKeyBytes)
	return privateKeyString
}

func (a *Account) Address() string {
	return crypto.PubkeyToAddress(a.privateKey.PublicKey).Hex()
}

func CreateAccount(name string) Account {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	account := Account{name, privateKey}
	return account
}

func AccountFromPrivateKey(name string, privateKeyHex string) Account {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatal(err)
	}
	account := Account{name, privateKey}
	return account
}

func (a *Account) SignMessage(message string) string {
	hash := DigestHash(message)
	signature, err := crypto.Sign(hash, a.privateKey)
	if err != nil {
		log.Fatal(err)
	}
	signature[crypto.RecoveryIDOffset] += 27 // hacky workaround.. don't even know why this is necessary
	return hexutil.Encode(signature)
}
