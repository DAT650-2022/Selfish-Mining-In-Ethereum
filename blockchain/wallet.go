package main

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	sha2562 "crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	_ "golang.org/x/crypto/ripemd160"
)

const (
	version            = byte(0x00)
	addressChecksumLen = 4
)

// newKeyPair creates a new cryptographic key pair
func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	privateKey, _ := ecdsa.GenerateKey(curve, rand.Reader)
	publicKeyByte := pubKeyToByte(privateKey.PublicKey)
	return *privateKey, publicKeyByte
}

// pubKeyToByte converts the ecdsa.PublicKey to a concatenation of its coordinates in bytes
// step 1 of: https://en.bitcoin.it/wiki/Technical_background_of_version_1_Bitcoin_addresses#How_to_create_Bitcoin_Address
func pubKeyToByte(pubkey ecdsa.PublicKey) []byte {
	return append(pubkey.X.Bytes(), pubkey.Y.Bytes()...)
}

// GetAddress returns address
// https://en.bitcoin.it/wiki/Technical_background_of_version_1_Bitcoin_addresses#How_to_create_Bitcoin_Address
func GetAddress(pubKeyBytes []byte) []byte {
	pubKeyHash := HashPubKey(pubKeyBytes)
	pubKeyHash = append([]byte{version}, pubKeyHash...)
	check := checksum(pubKeyHash)
	pubKeyHash = append(pubKeyHash, check...)
	address := Base58Encode(pubKeyHash)
	return address
}

// GetStringAddress returns address as string
func GetStringAddress(pubKeyBytes []byte) string {
	return string(GetAddress(pubKeyBytes))
}

// HashPubKey hashes public key
func HashPubKey(pubKey []byte) []byte {
	sha := sha2562.Sum256(pubKey)
	ripemd := crypto.RIPEMD160.New()
	ripemd.Write(sha[:])
	return ripemd.Sum(nil)
}

// GetPubKeyHashFromAddress returns the hash of the public key
// discarding the version and the checksum
// how it is stored: version + pubkeyhash + checksum
func GetPubKeyHashFromAddress(address string) []byte {
	base58 := Base58Decode([]byte(address))
	return base58[1 : len(base58)-addressChecksumLen]
}

// ValidateAddress check if an address is valid
func ValidateAddress(address string) bool {
	base58 := Base58Decode([]byte(address))
	check := base58[len(base58)-addressChecksumLen:]
	pubKeyHash := base58[:len(base58)-addressChecksumLen]
	return bytes.Equal(check, checksum(pubKeyHash))
}

// Checksum generates a checksum for a public key
func checksum(payload []byte) []byte {
	sha1 := sha2562.Sum256(payload)
	sha2 := sha2562.Sum256(sha1[:])
	return sha2[:addressChecksumLen]
}

func encodeKeyPair(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey) (string, string) {
	return encodePrivateKey(privateKey), encodePublicKey(publicKey)
}

func encodePrivateKey(privateKey *ecdsa.PrivateKey) string {
	x509Encoded, _ := x509.MarshalECPrivateKey(privateKey)
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})

	return string(pemEncoded)
}

func encodePublicKey(publicKey *ecdsa.PublicKey) string {
	x509EncodedPub, _ := x509.MarshalPKIXPublicKey(publicKey)
	pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})

	return string(pemEncodedPub)
}

func decodeKeyPair(pemEncoded string, pemEncodedPub string) (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
	return decodePrivateKey(pemEncoded), decodePublicKey(pemEncodedPub)
}

func decodePrivateKey(pemEncoded string) *ecdsa.PrivateKey {
	block, _ := pem.Decode([]byte(pemEncoded))
	privateKey, _ := x509.ParseECPrivateKey(block.Bytes)

	return privateKey
}

func decodePublicKey(pemEncodedPub string) *ecdsa.PublicKey {
	blockPub, _ := pem.Decode([]byte(pemEncodedPub))
	genericPubKey, _ := x509.ParsePKIXPublicKey(blockPub.Bytes)
	publicKey := genericPubKey.(*ecdsa.PublicKey) // cast to ecdsa

	return publicKey
}
