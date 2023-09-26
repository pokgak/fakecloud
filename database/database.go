package database

import "database/sql"

var db *sql.DB

func SetDB(database *sql.DB) {
	db = database
}

func GetDB() *sql.DB {
	return db
}
