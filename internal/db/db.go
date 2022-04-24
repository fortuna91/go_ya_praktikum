package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v4"
)

func Ping(dbConn *pgx.Conn) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := dbConn.Ping(ctx); err != nil {
		return false
	}
	return true
}

func Connect(dbAddress string) *pgx.Conn {
	/*dbConn, err := sql.Open("pgx", dbAddress)
	if err != nil {
		panic(err)
	}*/
	dbConn, err := pgx.Connect(context.Background(), dbAddress)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Set db connection...")
	return dbConn
}
