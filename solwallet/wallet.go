package main

import (
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/tyler-smith/go-bip39"
	"strconv"
	"strings"
)

const (
	HardenedOffset = 0x80000000
)

type HDKey struct {
	Key         []byte
	ChainCode   []byte
	Depth       uint8
	ChildNumber uint32
	ParentKey   []byte
}

type WalletInfo struct {
	Mnemonic   string
	PrivateKey string
	PublicKey  string
}

func NewMasterKey(seed []byte) (*HDKey, error) {
	hmac := hmac.New(sha512.New, []byte("ed25519 seed"))
	_, err := hmac.Write(seed)
	if err != nil {
		return nil, err
	}
	sum := hmac.Sum(nil)

	key := sum[:32]
	chainCode := sum[32:]

	return &HDKey{
		Key:         key,
		ChainCode:   chainCode,
		Depth:       0,
		ChildNumber: 0,
	}, nil
}

func (k *HDKey) Derive(path string) (*HDKey, error) {
	elements := strings.Split(path, "/")
	if elements[0] != "m" {
		return nil, fmt.Errorf("invalid path: must start with 'm'")
	}

	key := k
	for _, element := range elements[1:] {
		hardened := false
		if strings.HasSuffix(element, "'") {
			hardened = true
			element = strings.TrimRight(element, "'")
		}

		index, err := strconv.ParseUint(element, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid path element: %s", element)
		}

		if hardened {
			index += HardenedOffset
		}

		key, err = key.deriveChild(uint32(index))
		if err != nil {
			return nil, err
		}
	}

	return key, nil
}

func (k *HDKey) deriveChild(childNumber uint32) (*HDKey, error) {
	var data []byte
	if childNumber >= HardenedOffset {
		data = append([]byte{0x0}, k.Key...)
	} else {
		pubKey := ed25519.PublicKey(k.Key[32:])
		data = pubKey
	}

	childIndexBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(childIndexBytes, childNumber)
	data = append(data, childIndexBytes...)

	hmac := hmac.New(sha512.New, k.ChainCode)
	_, err := hmac.Write(data)
	if err != nil {
		return nil, err
	}
	sum := hmac.Sum(nil)

	childKey := sum[:32]
	childChainCode := sum[32:]

	return &HDKey{
		Key:         childKey,
		ChainCode:   childChainCode,
		Depth:       k.Depth + 1,
		ChildNumber: childNumber,
		ParentKey:   k.Key,
	}, nil
}

func generateWallet() (*WalletInfo, error) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return nil, err
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, err
	}

	seed := bip39.NewSeed(mnemonic, "")

	masterKey, err := NewMasterKey(seed)
	if err != nil {
		return nil, err
	}

	path := "m/44'/501'/0'/0'"
	derivedKey, err := masterKey.Derive(path)
	if err != nil {
		return nil, err
	}

	privateKey := ed25519.NewKeyFromSeed(derivedKey.Key)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	return &WalletInfo{
		Mnemonic:   mnemonic,
		PrivateKey: base58.Encode(privateKey),
		PublicKey:  base58.Encode(publicKey),
	}, nil
}
