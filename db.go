package main

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	err = createTopic("hall")
	if err != nil {
		log.Fatal(err)
	}
}

func createTopic(name string) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS ` + name + `(
			id    char(16)      NOT NULL,
			user  varchar(16)   NOT NULL,
			value varchar(2048) NOT NULL,
			time  datetime      NOT NULL)`)
	if err != nil {
		return err
	}
	return nil
}
