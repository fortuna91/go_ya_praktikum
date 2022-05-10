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

type dbStorage struct {
	dbConnection *sql.DB
}

func New(dbAddress string) *dbStorage {
	dbConn, err := sql.Open("pgx", dbAddress)
	if err != nil {
		panic(err)
	}
	return &dbStorage{
		dbConnection: dbConn,
	}
}

func (db *dbStorage) Ping(ctx context.Context) bool {
	if err := db.dbConnection.PingContext(ctx); err != nil {
		return false
	}
	return true
}

func (db *dbStorage) Create(ctx context.Context) {
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s ("+
		"id varchar(50) NOT NULL,"+
		"type varchar(50) NOT NULL,"+
		"delta bigint,"+
		"value float,"+
		"CONSTRAINT id_type PRIMARY KEY(id,type));", tableName)
	_, err := db.dbConnection.ExecContext(ctx, query)
	if err != nil {
		panic(err)
	}
}

func (db *dbStorage) Close() {
	db.dbConnection.Close()
}

func (db *dbStorage) StoreMetric(ctx context.Context, metric *metrics.Metric) error {
	if metric.MType == metrics.Gauge {
		return db.setGauge(ctx, metric.ID, metric.Value)
	} else if metric.MType == metrics.Counter {
		return db.updateCounter(ctx, metric.ID, *metric.Delta)
	}
	return fmt.Errorf("unknown metric type %s", metric.MType)
}

func (db *dbStorage) setGauge(ctx context.Context, id string, val *float64) error {
	_, err := db.dbConnection.ExecContext(ctx, "INSERT INTO metrics (id, type, value) VALUES ($1, $2, $3) ON CONFLICT ON CONSTRAINT id_type DO UPDATE SET value = $3;",
		id, metrics.Gauge, *val)
	if err != nil {
		return fmt.Errorf("couldn't set metric %s into DB: %s", id, err)
	}
	return nil
}

func (db *dbStorage) updateCounter(ctx context.Context, id string, val int64) error {

	// excluded.delta
	_, err := db.dbConnection.ExecContext(ctx, "INSERT INTO metrics (id, type, delta) VALUES ($1, $2, $3) ON CONFLICT ON CONSTRAINT id_type DO UPDATE SET delta = $3;",
		id, metrics.Counter, val)
	if err != nil {
		return fmt.Errorf("couldn't set metric %s into DB: %s", id, err)
	}
	return nil
}

func (db *dbStorage) Restore() map[string]*metrics.Metric {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.dbConnection.QueryContext(ctx, "SELECT * FROM metrics")
	if err != nil {
		log.Println("Couldn't get metrics for restore")
		return nil
	}

	log.Println("Restore metrics from DB...")
	restoredMetrics := make(map[string]*metrics.Metric)
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
		restoredMetrics[resMetric.ID] = &resMetric
	}
	err = rows.Err()
	if err != nil {
		return nil
	}
	return restoredMetrics
}

func (db *dbStorage) StoreBatchMetrics(ctx context.Context, metricList []metrics.Metric) error {
	tx, err := db.dbConnection.Begin()
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
	log.Println("Store metrics into DB...")
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// better use SQLx here

func Get(dbConn *sql.DB, id string, mType string) *metrics.Metric {
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
