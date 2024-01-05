package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func GenerateECDSAKeys() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

func SignECDSA(privateKey *ecdsa.PrivateKey, data []byte) ([]byte, error) {
	hash := sha256.Sum256(data)
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		return nil, err
	}
	signature := append(r.Bytes(), s.Bytes()...)
	return signature, nil
}

type User struct {
	UserUuid   string `json:"user_uuid"`
	PublicKey  []byte `json:"public_key_base64"`
	PrivateKey []byte `json:"private_key_base64"`
}

func ToBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

type RegisterRequest struct {
	PublicKeyBase64 string `json:"public_key_base64"`
}

type RegisterResponse struct {
	UserUuid string `json:"user_uuid"`
}

func GenUserData() ([]byte, []byte) {
	// gen public and private keys
	privateKey, err := GenerateECDSAKeys()
	if err != nil {
		panic(err)
	}
	// convert private key to base64
	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		panic(err)
	}
	// convert public key to base64
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		panic(err)
	}

	return privateKeyBytes, publicKeyBytes
}

func Register(user *User) int {
	// send request
	req_body := RegisterRequest{
		PublicKeyBase64: ToBase64(user.PublicKey),
	}
	req_body_json, err := json.Marshal(req_body)
	if err != nil {
		panic(err)
	}

	res, err := http.Post("http://localhost:30001/v1/register", "application/json", bytes.NewBuffer(req_body_json))
	if err != nil {
		log.Fatalln(err)
	}

	// get user uuid from response
	res_body_json, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var res_body RegisterResponse
	json.Unmarshal(res_body_json, &res_body)
	user.UserUuid = res_body.UserUuid
	return res.StatusCode
}

func SignData(user *User, data []byte) []byte {
	privateKey, _ := x509.ParseECPrivateKey(user.PrivateKey)
	signature, err := SignECDSA(privateKey, data)
	if err != nil {
		panic(err)
	}

	return signature
}

type ValidateRequest struct {
	UserUuid        string `json:"user_uuid"`
	DataBase64      string `json:"data_base64"`
	SignatureBase64 string `json:"signature_base64"`
}

func Validate(user *User, data []byte) int {
	// send request
	req_body := ValidateRequest{
		UserUuid:        user.UserUuid,
		DataBase64:      ToBase64(data),
		SignatureBase64: ToBase64(SignData(user, data)),
	}
	req_body_json, err := json.Marshal(req_body)
	if err != nil {
		panic(err)
	}

	res, err := http.Post("http://localhost:30001/v1/validate", "application/json", bytes.NewBuffer(req_body_json))
	if err != nil {
		log.Fatalln(err)
	}

	// check response
	return res.StatusCode
}

func ValidateBroken(user *User, data []byte) int {
	// send request
	req_body := ValidateRequest{
		UserUuid:        user.UserUuid,
		DataBase64:      ToBase64(append(data, []byte("123")...)),
		SignatureBase64: ToBase64(SignData(user, data)),
	}
	req_body_json, err := json.Marshal(req_body)
	if err != nil {
		panic(err)
	}

	res, err := http.Post("http://localhost:30001/v1/validate", "application/json", bytes.NewBuffer(req_body_json))
	if err != nil {
		log.Fatalln(err)
	}

	// check response
	return res.StatusCode
}

func main() {
	user := &User{}
	user.PrivateKey, user.PublicKey = GenUserData()

	if Register(user) != http.StatusOK {
		fmt.Println("register test failed")
		return
	}

	data := []byte("my data")
	if Validate(user, data) != http.StatusOK {
		fmt.Println("validation test failed")
		return
	}

	data = []byte("other data")
	if ValidateBroken(user, data) != http.StatusBadRequest {
		fmt.Println("broken validation test failed")
		return
	}

	fmt.Println("ok")
}
