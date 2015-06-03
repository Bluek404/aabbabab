package main

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func initDB() error {
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	err = createTopic("hall")
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS lastID (id char(8))`)
	if err != nil {
		return err
	}
	row := db.QueryRow(`SELECT COUNT(*) FROM lastID`)
	var l int
	err = row.Scan(&l)
	if err != nil {
		return err
	}
	if l == 0 {
		_, err = db.Exec(`INSERT INTO lastID VALUES ('00000000')`)
		if err != nil {
			return err
		}
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS topics (
			id       char(8)     NOT NULL,
			title    varchar(50) NOT NULL,
			author   varchar(16) NOT NULL,
			modified datetime    NOT NULL)`)
	if err != nil {
		return err
	}

	return nil
}

func createTopic(name string) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS t_` + name + `(
			id    char(16)      NOT NULL,
			user  varchar(16)   NOT NULL,
			value varchar(2048) NOT NULL,
			time  datetime      NOT NULL)`)
	if err != nil {
		return err
	}
	return nil
}
