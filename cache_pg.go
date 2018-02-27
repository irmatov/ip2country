package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// this cache assumes following database table structure:
//
// CREATE TABLE ip_country_cache (
//     ip VARCHAR NOT NULL PRIMARY KEY,
//     country VARCHAR NOT NULL DEFAULT '',
//     expires_at TIMESTAMP WITH TIME ZONE NOT NULL
// );
type pgCache struct {
	db *sql.DB
}

func newPgCache(params interface{}) (countryCache, error) {
	dsn, ok := params.(string)
	if !ok {
		return nil, fmt.Errorf("bad value for CacheParameters: %v", params)
	}
	db, err := sql.Open("postgres", dsn)
	return pgCache{db}, err
}

func (c pgCache) Put(ip, country string, expires time.Time) {
	tx, err := c.db.Begin()
	if err != nil {
		log.Printf("error on transaction begin: %v", err)
		return
	}
	r, err := tx.Exec(`UPDATE ip_country_cache SET country = $1, expires_at = $2 WHERE ip = $3`, country, expires, ip)
	if err != nil {
		log.Printf("error on update: %v", err)
		tx.Rollback()
		return
	}
	n, err := r.RowsAffected()
	if err != nil {
		log.Printf("error while checking affected rows: %v", err)
		tx.Rollback()
		return
	}
	if n > 0 {
		tx.Commit()
		return
	}
	_, err = tx.Exec(`INSERT INTO ip_country_cache (ip, country, expires_at) values ($1, $2, $3)`, ip, country, expires)
	if err != nil {
		log.Printf("error on insert: %v", err)
		tx.Rollback()
		return
	}
	tx.Commit()
}

func (c pgCache) Get(ip string) (string, bool) {
	row := c.db.QueryRow(`SELECT country FROM ip_country_cache WHERE ip = $1 AND expires_at > NOW()`, ip)
	var country string
	if err := row.Scan(&country); err != nil {
		if err != sql.ErrNoRows {
			log.Printf("error on select from database: %v", err)
		}
		return "", false
	}
	return country, true
}
