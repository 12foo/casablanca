package main

import (
	"database/sql"
	"golang.org/x/crypto/bcrypt"
)

func createTestData(dbpath, user, password, email string) error {
	odb, err := sql.Open("sqlite3", dbpath)
	if err != nil {
		return err
	}
	if _, err := odb.Exec("CREATE TABLE users (username text, password blob, email text)"); err != nil {
		return err
	}
	pw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if _, err := odb.Exec("INSERT INTO users VALUES (?, ?, ?)", user, pw, email); err != nil {
		return err
	}
	return nil
}
