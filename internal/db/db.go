package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/fortuna91/go_ya_praktikum/internal/metrics"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

const (
	tableName = "metrics"
	dbName    = "storage"
)

// if we will have more than one sql query per request,move dbConn into handler

func Ping(dbAddress string) bool {
	dbConn := connect(dbAddress, true)
	defer dbConn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := dbConn.PingContext(ctx); err != nil {
		return false
	}
	return true
}

func connect(dbAddress string, connectToDB bool) *sql.DB {
	if connectToDB {
		dbAddress = dbAddress + "/" + dbName
	}
	dbConn, err := sql.Open("pgx", dbAddress)
	if err != nil {
		panic(err)
	}
	/*dbConn, err := pgx.connect(context.Background(), dbAddress)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}*/
	return dbConn
}

func CreateDB(dbAddress string) {
	dbConn := connect(dbAddress, false)
	defer dbConn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	/*_, err := dbConn.ExecContext(ctx, "SELECT 'CREATE DATABASE storagw' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'storage')\\gexec")
	if err != nil {
		panic(err)
	}*/
	query := fmt.Sprintf("CREATE DATABASE %s", dbName)
	_, err := dbConn.ExecContext(ctx, query)
	if err != nil {
		fmt.Printf("Database %s exists: %s\n", dbName, err)
	}
}

func CreateTable(dbAddress string) {
	dbConn := connect(dbAddress, true)
	defer dbConn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s ("+
		"id varchar(50) NOT NULL,"+
		"type varchar(50) NOT NULL,"+
		"delta integer,"+
		"value float,"+
		"CONSTRAINT id_type PRIMARY KEY(id,type));", tableName)
	_, err := dbConn.ExecContext(ctx, query)
	if err != nil {
		panic(err)
	}
}

func SetGauge(dbAddress string, id string, val *float64) bool {
	dbConn := connect(dbAddress, true)
	defer dbConn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := dbConn.ExecContext(ctx, "INSERT INTO metrics (id, type, value) VALUES ($1, $2, $3) ON CONFLICT ON CONSTRAINT id_type DO UPDATE SET value = $3;",
		id, metrics.Gauge, *val)
	if err != nil {
		fmt.Printf("Couldn't set metric %s into DB: %s\n", id, err)
		return false
	}
	return true
}

func UpdateCounter(dbAddress string, id string, val int64) bool {
	dbConn := connect(dbAddress, true)
	defer dbConn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := dbConn.ExecContext(ctx, "INSERT INTO metrics (id, type, delta) VALUES ($1, $2, $3) ON CONFLICT ON CONSTRAINT id_type DO UPDATE SET delta = EXCLUDED.delta + $3;",
		id, metrics.Counter, val)
	if err != nil {
		fmt.Printf("Couldn't set metric %s into DB: %s\n", id, err)
		return false
	}
	return true
}

func Get(dbAddress string, id string, mType string) *metrics.Metric {
	dbConn := connect(dbAddress, true)
	defer dbConn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resMetric := metrics.Metric{}
	var delta sql.NullInt64
	var value sql.NullFloat64

	err := dbConn.QueryRowContext(ctx, "SELECT * FROM metrics WHERE id=$1 AND type=$2", id, mType).Scan(&resMetric.ID, &resMetric.MType, &delta, &value)
	if err != nil {
		fmt.Printf("Couldn't get metric %s from DB: %s\n", id, err)
		return nil
	}
	if delta.Valid {
		resMetric.Delta = &delta.Int64
	}
	if value.Valid {
		resMetric.Value = &value.Float64
	}
	return &resMetric
}
