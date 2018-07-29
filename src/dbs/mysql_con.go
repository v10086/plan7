package dbs

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

var (
	db *sql.DB
)

func init() {
	var err error
	var dsn string = "root:abc1688@tcp(127.0.0.1:3306)/goapp?charset=utf8"
	db, err = sql.Open("mysql", dsn)
	if err == nil {
		db.SetMaxOpenConns(2000)
		db.SetMaxIdleConns(100)
	} else {
		panic(err)
	}
}

func GetCon() *sql.DB {
	return db
}
