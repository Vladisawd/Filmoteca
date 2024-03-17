package main

import (
	"database/sql"
	"fmt"
)

func connect(conf setting) *sql.DB {
	var err error

	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable", conf.PgHost, conf.PgPort, conf.PgBase, conf.PgUser, conf.PgPassword))
	if err != nil {
		panic(fmt.Sprintf("Нет коннекта %s", err))
	}

	if err = db.Ping(); err != nil {
		panic(fmt.Sprintf("Нет коннекта к БД %s", err))
	}

	return db
}
