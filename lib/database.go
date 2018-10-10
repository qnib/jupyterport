package qniblib // import "github.com/qnib/jupyterport/lib"

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)


type Database struct {
	db *sql.DB
}


func NewDatabase(fname string) Database {
	d, err := sql.Open("sqlite3", fname)
	checkErr(err)
	return Database{db: d}
}


func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}