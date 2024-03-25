package handlers

import (
	"blockchain-voting-validator/internal/database"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
)

type ValidateHandler struct {
	Postgres *database.Postgres
}

func NewValidateHandler(postgres *database.Postgres) *ValidateHandler {
	return &ValidateHandler{
		Postgres: postgres,
	}
}

type ValidateRequest struct {
	UserUuid        string `json:"user_uuid"`
	VotingUuid      string `json:"voting_uuid,omitempty"`
	DataBase64      string `json:"data_base64"`
	SignatureBase64 string `json:"signature_base64"`
}

// Проверка подписи ECDSA
func VerifyECDSASignature(publicKey *ecdsa.PublicKey, data, signature []byte) bool {
	hash := sha256.Sum256(data)
	r := new(big.Int).SetBytes(signature[:len(signature)/2])
	s := new(big.Int).SetBytes(signature[len(signature)/2:])
	return ecdsa.Verify(publicKey, hash[:], r, s)
}

func ConvertPublicKeyBase64ToKey(publicKeyBase64 string) *ecdsa.PublicKey {
	publicKey, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return nil
	}
	key_, err := x509.ParsePKIXPublicKey(publicKey)
	if err != nil {
		return nil
	}
	return key_.(*ecdsa.PublicKey)
}

func Verify(publicKeyBase64, dataBase64, signatureBase64 string) bool {
	publicKey := ConvertPublicKeyBase64ToKey(publicKeyBase64)
	if publicKey == nil {
		return false
	}
	data, err := base64.StdEncoding.DecodeString(dataBase64)
	if err != nil {
		return false
	}
	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return false
	}
	return VerifyECDSASignature(publicKey, data, signature)
}

func (handler *ValidateHandler) SelectPublicKeyFromDatabase(userUuid string) (string, error) {
	var publicKeyBase64 string
	connection := handler.Postgres.Connect()
	err := connection.QueryRow(
		"SELECT public_key_base64 FROM validator.users WHERE uuid = $1;",
		userUuid,
	).Scan(&publicKeyBase64)
	return publicKeyBase64, err
}

func (handler *ValidateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotFound)
	}

	var body ValidateRequest

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(bodyBytes, &body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	publicKeyBase64, err := handler.SelectPublicKeyFromDatabase(body.UserUuid)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if Verify(publicKeyBase64, body.DataBase64, body.SignatureBase64) {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}
