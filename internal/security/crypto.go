package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

// LoadPublicKey загружает RSA публичный ключ из X.509 сертификата в формате PEM
func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	publicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("certificate does not contain RSA public key")
	}

	return publicKey, nil
}

// LoadPrivateKey загружает RSA приватный ключ из файла в формате PEM
func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		parsedKey, errPKCS8 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if errPKCS8 != nil {
			return nil, err
		}

		var ok bool
		privateKey, ok = parsedKey.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("private key is not RSA")
		}
	}

	return privateKey, nil
}

// Encrypt шифрует данные RSA-OAEP с SHA-256, разбивая длинные сообщения на части
func Encrypt(plaintext []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	if len(plaintext) == 0 {
		return plaintext, nil
	}

	chunkSize := publicKey.Size() - 2*sha256.Size - 2
	encrypted := make([]byte, 0, ((len(plaintext)+chunkSize-1)/chunkSize)*publicKey.Size())

	for start := 0; start < len(plaintext); start += chunkSize {
		end := start + chunkSize
		if end > len(plaintext) {
			end = len(plaintext)
		}

		chunk, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, plaintext[start:end], nil)
		if err != nil {
			return nil, err
		}

		encrypted = append(encrypted, chunk...)
	}

	return encrypted, nil
}

// Decrypt расшифровывает данные, зашифрованные функцией Encrypt
func Decrypt(ciphertext []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	if len(ciphertext) == 0 {
		return ciphertext, nil
	}

	chunkSize := privateKey.PublicKey.Size()
	if len(ciphertext)%chunkSize != 0 {
		return nil, errors.New("invalid ciphertext length")
	}

	decrypted := make([]byte, 0, len(ciphertext))

	for start := 0; start < len(ciphertext); start += chunkSize {
		chunk, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, ciphertext[start:start+chunkSize], nil)
		if err != nil {
			return nil, err
		}

		decrypted = append(decrypted, chunk...)
	}

	return decrypted, nil
}
