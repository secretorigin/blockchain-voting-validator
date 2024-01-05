package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
)

// Генерация пары ключей ECDSA
func GenerateECDSAKeys() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

// Создание подписи ECDSA
func SignECDSA(privateKey *ecdsa.PrivateKey, data []byte) ([]byte, error) {
	hash := sha256.Sum256(data)
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		return nil, err
	}
	signature := append(r.Bytes(), s.Bytes()...)
	return signature, nil
}

// Проверка подписи ECDSA
func VerifyECDSASignature(publicKey *ecdsa.PublicKey, data, signature []byte) bool {
	hash := sha256.Sum256(data)
	r := new(big.Int).SetBytes(signature[:len(signature)/2])
	s := new(big.Int).SetBytes(signature[len(signature)/2:])
	return ecdsa.Verify(publicKey, hash[:], r, s)
}

func main() {
	// Генерация ключей
	privateKey, err := GenerateECDSAKeys()
	if err != nil {
		panic(err)
	}

	// Экспорт открытого ключа в формате PEM
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		panic(err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	publicKeyBase64 := base64.StdEncoding.EncodeToString([]byte(publicKeyBytes))
	fmt.Println(publicKeyBase64)
	fmt.Println(string(publicKeyPEM))

	// Экспорт закрытого ключа в формате PEM
	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		panic(err)
	}
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	privateKeyBase64 := base64.StdEncoding.EncodeToString([]byte(privateKeyBytes))
	fmt.Println(privateKeyBase64)
	fmt.Println(string(privateKeyPEM))

	// Сохранение ключей в файлы
	err = os.WriteFile("public.pem", publicKeyPEM, 0644)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("private.pem", privateKeyPEM, 0644)
	if err != nil {
		panic(err)
	}

	// Создание подписи
	data := []byte("Пример данных для подписи")
	signature, err := SignECDSA(privateKey, data)
	if err != nil {
		panic(err)
	}

	// Проверка подписи
	valid := VerifyECDSASignature(&privateKey.PublicKey, data, signature)
	if !valid {
		panic("Подпись недействительна")
	} else {
		println("Подпись успешно проверена")
	}
}
