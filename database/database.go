package database

import (
	"github.com/jmoiron/sqlx"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var DB *sqlx.DB

const DATABASE_URL string = "postgres://pnunn:m4R13cat23*@192.168.44.109:5432/projectmotor_dev?search_path=public&sslmode=disable"

func OpenDB() (*sqlx.DB, error) {
	conn, err := sqlx.Connect("pgx", DATABASE_URL)
	if err != nil {
		return conn, err
	}
	DB = conn
	return conn, nil
}

func CloseDB() error {
	return DB.Close()
}

// NOTE ->> only for testing, remove after actual interactions with database
func SetupDB() error {
	_, err := DB.Exec(`create table if not exists messages ("id" serial PRIMARY KEY, "message" text NOT NULL);`)
	if err != nil {
		return err
	}
	_, err = DB.Exec(`insert into messages (message) values ('Hello, world!');`)
	if err != nil {
		return err
	}
	return nil
}

func GetMessage() (string, error) {
	var message string
	err := DB.QueryRow("select message from messages limit 1;").Scan(&message)
	if err != nil {
		return "", err
	}
	return message, nil
}

// <<- NOTE
