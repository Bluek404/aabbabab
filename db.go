package main

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

var insTopicStmt, upLastIdStmt, upModTimeStmt, getTopicListStmt *sql.Stmt

func initStmt() (err error) {
	insTopicStmt, err = db.Prepare(`
		INSERT INTO topics (id, title, author, time, modified)
		VALUES             (?,  ?,     ?,      ?,    ?       )`)
	if err != nil {
		return err
	}

	upLastIdStmt, err = db.Prepare(`UPDATE lastID SET id = ? WHERE id = ?`)
	if err != nil {
		return err
	}

	upModTimeStmt, err = db.Prepare(`UPDATE topics SET modified = ? WHERE id = ?`)
	if err != nil {
		return err
	}

	getTopicListStmt, err = db.Prepare(`SELECT id, title, author, time FROM topics ORDER BY modified DESC`)
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

	var l int
	err = db.QueryRow(`SELECT COUNT(*) FROM lastID`).Scan(&l)
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
			time     datetime    NOT NULL,
			modified datetime    NOT NULL)`)
	if err != nil {
		return err
	}

	return initStmt()
}
