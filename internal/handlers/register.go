package handlers

import (
	"blockchain-voting-validator/internal/database"
	"encoding/json"
	"io"
	"net/http"
)

type RegisterHandler struct {
	Postgres *database.Postgres
}

func NewRegisterHandler(postgres *database.Postgres) *RegisterHandler {
	return &RegisterHandler{
		Postgres: postgres,
	}
}

type RegisterRequest struct {
	PublicKeyBase64 string `json:"public_key_base64"`
}

type RegisterResponse struct {
	UserUuid string `json:"user_uuid"`
}

func (handler *RegisterHandler) InsertUserInDatabase(publicKeyBase64 string) (string, error) {
	var userUuid string
	connection := handler.Postgres.Connect()
	err := connection.QueryRow(
		"INSERT INTO validator.users (public_key_base64) VALUES ($1) RETURNING uuid;",
		publicKeyBase64,
	).Scan(&userUuid)
	return userUuid, err
}

func (handler *RegisterHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotFound)
	}

	var req_body RegisterRequest
	var res_body RegisterResponse

	req_body_bytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(req_body_bytes, &req_body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res_body.UserUuid, err = handler.InsertUserInDatabase(req_body.PublicKeyBase64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res_body_bytes, err := json.Marshal(res_body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(res_body_bytes)
}
