package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Postgres struct {
	cfg *PostgresConfig
}

func NewPostgres(cfg *PostgresConfig) *Postgres {
	return &Postgres{
		cfg: cfg,
	}
}

type PostgresConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Dbname   string `json:"dbname"`
}

func (d *Postgres) Connect() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		d.cfg.Host, d.cfg.Port, d.cfg.User, d.cfg.Password, d.cfg.Dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil
	}

	return db
}
