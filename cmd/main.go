package main

import (
	"blockchain-voting-validator/internal/database"
	"blockchain-voting-validator/internal/handlers"
	"net/http"
)

func main() {
	postgresConfig := &database.PostgresConfig{
		Host:     "postgres",
		Port:     5432,
		User:     "validator",
		Password: "Validator12345",
		Dbname:   "validator_db",
	}
	postgres := database.NewPostgres(postgresConfig)

	http.HandleFunc("/v1/register", handlers.NewRegisterHandler(postgres).Handle)
	http.HandleFunc("/v1/validate", handlers.NewValidateHandler(postgres).Handle)

	http.ListenAndServe(":30001", nil)
}
