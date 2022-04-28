package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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

func CreateTable(dbAddress string) {
	dbConn := connect(dbAddress)
	defer dbConn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s ("+
		"id varchar(50) NOT NULL,"+
		"type varchar(50) NOT NULL,"+
		"delta bigint,"+
		"value float,"+
		"CONSTRAINT id_type PRIMARY KEY(id,type));", tableName)
	_, err := dbConn.ExecContext(ctx, query)
	if err != nil {
		panic(err)
	}
}

func SetGauge(dbAddress string, id string, val *float64) error {
	dbConn := connect(dbAddress)
	defer dbConn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := dbConn.ExecContext(ctx, "INSERT INTO metrics (id, type, value) VALUES ($1, $2, $3) ON CONFLICT ON CONSTRAINT id_type DO UPDATE SET value = $3;",
		id, metrics.Gauge, *val)
	if err != nil {
		return fmt.Errorf("couldn't set metric %s into DB: %s", id, err)
	}
	return nil
}

func UpdateCounter(dbAddress string, id string, val int64) error {
	dbConn := connect(dbAddress)
	defer dbConn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// excluded.delta
	_, err := dbConn.ExecContext(ctx, "INSERT INTO metrics (id, type, delta) VALUES ($1, $2, $3) ON CONFLICT ON CONSTRAINT id_type DO UPDATE SET delta = $3;",
		id, metrics.Counter, val)
	if err != nil {
		return fmt.Errorf("couldn't set metric %s into DB: %s", id, err)
	}
	return nil
}

func Restore(dbAddress string) map[string]*metrics.Metric {
	dbConn := connect(dbAddress)
	defer dbConn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := dbConn.QueryContext(ctx, "SELECT * FROM metrics")
	if err != nil {
		log.Println("Couldn't get metrics for restore")
		return nil
	}
	restoreMetrics := make(map[string]*metrics.Metric)
	for rows.Next() {
		resMetric := metrics.Metric{}
		var delta sql.NullInt64
		var value sql.NullFloat64
		err = rows.Scan(&resMetric.ID, &resMetric.MType, &delta, &value)
		if err != nil {
			log.Printf("Couldn't get metric %s from DB: %s\n", resMetric.ID, err)
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

// SQLx

func Get(dbAddress string, id string, mType string) *metrics.Metric {
	dbConn := connect(dbAddress)
	defer dbConn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resMetric := metrics.Metric{}
	var delta sql.NullInt64
	var value sql.NullFloat64

	err := dbConn.QueryRowContext(ctx, "SELECT * FROM metrics WHERE id=$1 AND type=$2", id, mType).Scan(&resMetric.ID, &resMetric.MType, &delta, &value)
	if err != nil {
		log.Printf("Couldn't get metric %s from DB: %s\n", id, err)
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

func SetBatchMetrics(dbAddress string, metricList []metrics.Metric) error {
	dbConn := connect(dbAddress)
	defer dbConn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO metrics (id, type, delta, value) VALUES ($1, $2, $3, $4) ON CONFLICT ON CONSTRAINT id_type DO UPDATE SET value = excluded.value, delta = excluded.delta")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, m := range metricList {
		if _, err = stmt.ExecContext(ctx, m.ID, m.MType, m.Delta, m.Value); err != nil {
			if errRollback := tx.Rollback(); errRollback != nil {
				log.Fatalf("update drivers: unable to rollback: %v", errRollback)
			}
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
