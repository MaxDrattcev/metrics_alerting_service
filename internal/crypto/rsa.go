package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

// LoadPublicKey загружает публичный RSA-ключ из PEM-файла.
func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	block, err := readPEMBlock(path)
	if err != nil {
		return nil, err
	}
	switch block.Type {
	case "PUBLIC KEY":
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		key, ok := pub.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("not RSA public key")
		}
		return key, nil
	case "RSA PUBLIC KEY":
		return x509.ParsePKCS1PublicKey(block.Bytes)
	case "CERTIFICATE":
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}
		key, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("certificate does not contain a RSA public key")
		}
		return key, nil
	default:
		return nil, fmt.Errorf("unsupported public key type: %s", block.Type)
	}
}

// LoadPrivateKey загружает приватный RSA-ключ из PEM-файла.
func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	block, err := readPEMBlock(path)
	if err != nil {
		return nil, err
	}
	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("not RSA private key")
		}
		return rsaKey, nil
	default:
		return nil, fmt.Errorf("unsupported private key type: %s", block.Type)
	}
}

// oaepChunkSize — максимальный размер блока для RSA-OAEP (SHA-256).
func oaepChunkSize(keySize int) int {
	return keySize - 2*sha256.Size - 2
}

// Encrypt шифрует данные публичным ключом (RSA-OAEP + SHA-256, по блокам).
func Encrypt(pub *rsa.PublicKey, data []byte) ([]byte, error) {
	if pub == nil {
		return nil, errors.New("public key is nil")
	}
	chunkSize := oaepChunkSize(pub.Size())
	if chunkSize <= 0 {
		return nil, errors.New("invalid public key size")
	}
	out := make([]byte, 0, ((len(data)+chunkSize-1)/chunkSize)*pub.Size())
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		part, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pub, data[i:end], nil)
		if err != nil {
			return nil, err
		}
		out = append(out, part...)
	}
	return out, nil
}

// Decrypt расшифровывает данные приватным ключом (RSA-OAEP + SHA-256, по блокам).
func Decrypt(priv *rsa.PrivateKey, data []byte) ([]byte, error) {
	if priv == nil {
		return nil, errors.New("private key is nil")
	}
	chunkSize := priv.Size()
	if len(data) == 0 || len(data)%chunkSize != 0 {
		return nil, errors.New("invalid encrypted data length")
	}
	out := make([]byte, 0, len(data))
	for i := 0; i < len(data); i += chunkSize {
		part, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, priv, data[i:i+chunkSize], nil)
		if err != nil {
			return nil, err
		}
		out = append(out, part...)
	}
	return out, nil
}

func readPEMBlock(path string) (*pem.Block, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read key file: %w", err)
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	return block, nil
}
