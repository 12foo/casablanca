package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/ziutek/mymysql/native"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

type AuthenticatedUser struct {
	Username string
	Extra    map[string]interface{}
}

type Backend interface {
	Authenticate(user, password string) *AuthenticatedUser
}

type DatabaseBackend struct {
	db      *sql.DB
	query   *sql.Stmt
	extrasV []string
}

func (db *DatabaseBackend) Authenticate(user, password string) *AuthenticatedUser {
	r := db.query.QueryRow(user)
	row := []interface{}{}
	if err := r.Scan(&row); err != nil {
		return nil
	}
	if len(row) < 2 {
		return nil
	}
	pw := row[1].([]byte)
	if bcrypt.CompareHashAndPassword(pw, []byte(password)) == nil {
		au := AuthenticatedUser{Username: row[0].(string)}
		emap := make(map[string]interface{}, len(db.extrasV))
		for i, e := range row[2:] {
			emap[db.extrasV[i]] = e
		}
		au.Extra = emap
		return &au
	}
	return nil
}

func NewDatabaseBackend(config map[string]interface{}) (*DatabaseBackend, error) {
	db := DatabaseBackend{}
	odb, err := sql.Open(config["driver"].(string), config["connection"].(string))
	if err != nil {
		return nil, err
	}
	db.db = odb

	tableName := config["table"].(string)
	userCol := config["username_col"].(string)
	pwCol := config["password_col"].(string)
	extrasK := []string{}
	db.extrasV = []string{}
	if _, ok := config["mapping"]; ok {
		tmp := config["mapping"].(map[string]interface{})
		for k, v := range tmp {
			extrasK = append(extrasK, k)
			db.extrasV = append(db.extrasV, v.(string))
		}
	}

	// TODO properly quote column names and table name
	q := fmt.Sprintf("SELECT %s, %s", userCol, pwCol)
	if len(extrasK) > 0 {
		q = q + ", " + strings.Join(extrasK, ",")
	}
	q = q + fmt.Sprintf(" FROM %s", tableName)
	q = q + fmt.Sprintf(" WHERE %s = ?", userCol)

	stmt, err := db.db.Prepare(q)
	if err != nil {
		return nil, err
	}
	db.query = stmt

	return &db, nil
}
