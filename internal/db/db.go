package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

func Ping(db *sql.DB) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return false
	}
	return true
}

func Connect(dbAddress string) *sql.DB {
	dbConn, err := sql.Open("postgres", dbAddress)
	if err != nil {
		panic(err)
	}
	fmt.Println("Set db connection...")
	return dbConn
}
