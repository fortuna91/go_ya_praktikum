package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/fortuna91/go_ya_praktikum/internal/metrics"
)

const tableName = "metrics"

// if we will have more than one sql query per request,move dbConn into handler

func Ping(dbAddress string) bool {
	dbConn := connect(dbAddress)
	if dbConn == nil {
		return false
	}
	defer dbConn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := dbConn.PingContext(ctx); err != nil {
		return false
	}
	return true
}

func connect(dbAddress string) *sql.DB {
	fmt.Println("dbAddress", dbAddress)
	dbConn, err := sql.Open("pgx", dbAddress)
	if err != nil {
		// panic(err)
		fmt.Printf("Unable to connect to database: %v\n", err)
		return nil
	}
	/*dbConn, err := pgx.connect(context.Background(), dbAddress)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}*/
	return dbConn
}

func CreateTable(dbAddress string) {
	dbConn := connect(dbAddress)
	if dbConn == nil {
		return
	}
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
	dbConn := connect(dbAddress)
	if dbConn == nil {
		return false
	}
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
	dbConn := connect(dbAddress)
	if dbConn == nil {
		return false
	}
	defer dbConn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := dbConn.ExecContext(ctx, "INSERT INTO metrics (id, type, delta) VALUES ($1, $2, $3) ON CONFLICT ON CONSTRAINT id_type DO UPDATE SET delta = $3;",
		id, metrics.Counter, val)
	if err != nil {
		fmt.Printf("Couldn't set metric %s into DB: %s\n", id, err)
		return false
	}
	return true
}

func StoreMetrics(handlerMetrics map[string]*metrics.Metric, dbAddress string) {
	dbConn := connect(dbAddress)
	if dbConn == nil {
		return
	}
	defer dbConn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, m := range handlerMetrics {
		_, err := dbConn.ExecContext(ctx, "INSERT INTO metrics (id, type, delta, value) VALUES ($1, $2, $3, $4) ON CONFLICT ON CONSTRAINT id_type DO UPDATE SET delta = $3, value = $4;",
			m.ID, m.MType, m.Delta, m.Value)
		if err != nil {
			fmt.Printf("Couldn't set metric %s into DB: %s\n", m.ID, err)
			return
		}
	}
}

func Restore(dbAddress string) map[string]*metrics.Metric {
	dbConn := connect(dbAddress)
	if dbConn == nil {
		return nil
	}
	defer dbConn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := dbConn.QueryContext(ctx, "SELECT * FROM metrics")
	if err != nil {
		return nil
	}
	restoreMetrics := make(map[string]*metrics.Metric)
	for rows.Next() {
		resMetric := metrics.Metric{}
		var delta sql.NullInt64
		var value sql.NullFloat64
		err = rows.Scan(&resMetric.ID, &resMetric.MType, &delta, &value)
		if err != nil {
			fmt.Printf("Couldn't get metric %s from DB: %s\n", resMetric.ID, err)
			return nil
		}
		if delta.Valid {
			resMetric.Delta = &delta.Int64
		}
		if value.Valid {
			resMetric.Value = &value.Float64
		}
		restoreMetrics[resMetric.ID] = &resMetric
	}
	err = rows.Err()
	if err != nil {
		return nil
	}
	return restoreMetrics
}

func Get(dbAddress string, id string, mType string) *metrics.Metric {
	dbConn := connect(dbAddress)
	if dbConn == nil {
		return nil
	}
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
