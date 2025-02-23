package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"order-system/pkg/infra/config"

	_ "github.com/go-sql-driver/mysql"
)

// Database represents a database connection
type Database interface {
	// Transaction executes a function within a transaction
	Transaction(ctx context.Context, fn func(Transaction) error) error

	// Exec executes a query without returning any rows
	Exec(ctx context.Context, query string, args ...interface{}) (*Result, error)

	// Query executes a query that returns rows
	Query(ctx context.Context, query string, args ...interface{}) ([]Row, error)

	// QueryRow executes a query that returns a single row
	QueryRow(ctx context.Context, query string, args ...interface{}) Row

	// Stats returns database statistics
	Stats() Stats

	// Close closes the database connection
	Close() error
}

// db implements the Database interface
type db struct {
	*sql.DB
	config *config.Config
}

// New creates a new database connection
func New(cfg *config.Config) (Database, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Database,
	)

	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, &Error{
			Operation: "open",
			Err:       err,
		}
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.MaxLifetime)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, &Error{
			Operation: "ping",
			Err:       err,
		}
	}

	return &db{
		DB:     sqlDB,
		config: cfg,
	}, nil
}

// Transaction executes a function within a transaction
func (d *db) Transaction(ctx context.Context, fn func(Transaction) error) error {
	tx, err := d.DB.BeginTx(ctx, nil)
	if err != nil {
		return &Error{
			Operation: "begin_transaction",
			Err:       err,
		}
	}

	// Create transaction wrapper
	txWrapper := &transaction{
		Tx: tx,
	}

	// Execute function
	if err := fn(txWrapper); err != nil {
		// Rollback on error
		if rbErr := tx.Rollback(); rbErr != nil {
			return &Error{
				Operation: "rollback",
				Err:       fmt.Errorf("rollback failed: %v (original error: %v)", rbErr, err),
			}
		}
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return &Error{
			Operation: "commit",
			Err:       err,
		}
	}

	return nil
}

// Exec executes a query without returning any rows
func (d *db) Exec(ctx context.Context, query string, args ...interface{}) (*Result, error) {
	result, err := d.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, &Error{
			Operation: "exec",
			Query:     query,
			Err:       err,
		}
	}

	lastInsertId, err := result.LastInsertId()
	if err != nil {
		return nil, &Error{
			Operation: "last_insert_id",
			Query:     query,
			Err:       err,
		}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, &Error{
			Operation: "rows_affected",
			Query:     query,
			Err:       err,
		}
	}

	return &Result{
		LastInsertId: lastInsertId,
		RowsAffected: rowsAffected,
	}, nil
}

// Query executes a query that returns rows
func (d *db) Query(ctx context.Context, query string, args ...interface{}) ([]Row, error) {
	rows, err := d.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, &Error{
			Operation: "query",
			Query:     query,
			Err:       err,
		}
	}
	defer rows.Close()

	var result []Row
	for rows.Next() {
		result = append(result, rows)
	}

	if err := rows.Err(); err != nil {
		return nil, &Error{
			Operation: "scan",
			Query:     query,
			Err:       err,
		}
	}

	return result, nil
}

// QueryRow executes a query that returns a single row
func (d *db) QueryRow(ctx context.Context, query string, args ...interface{}) Row {
	return d.DB.QueryRowContext(ctx, query, args...)
}

// Stats returns database statistics
func (d *db) Stats() Stats {
	stats := d.DB.Stats()
	return Stats{
		OpenConnections: stats.OpenConnections,
		InUse:           stats.InUse,
		Idle:            stats.Idle,
		WaitCount:       stats.WaitCount,
		WaitDuration:    stats.WaitDuration,
		MaxIdleTime:     time.Duration(stats.MaxIdleTimeClosed),
	}
}

// transaction implements the Transaction interface
type transaction struct {
	*sql.Tx
}

func (t *transaction) Exec(ctx context.Context, query string, args ...interface{}) (*Result, error) {
	result, err := t.Tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, &Error{
			Operation: "exec",
			Query:     query,
			Err:       err,
		}
	}

	lastInsertId, err := result.LastInsertId()
	if err != nil {
		return nil, &Error{
			Operation: "last_insert_id",
			Query:     query,
			Err:       err,
		}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, &Error{
			Operation: "rows_affected",
			Query:     query,
			Err:       err,
		}
	}

	return &Result{
		LastInsertId: lastInsertId,
		RowsAffected: rowsAffected,
	}, nil
}

func (t *transaction) Query(ctx context.Context, query string, args ...interface{}) ([]Row, error) {
	rows, err := t.Tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, &Error{
			Operation: "query",
			Query:     query,
			Err:       err,
		}
	}
	defer rows.Close()

	var result []Row
	for rows.Next() {
		result = append(result, rows)
	}

	if err := rows.Err(); err != nil {
		return nil, &Error{
			Operation: "scan",
			Query:     query,
			Err:       err,
		}
	}

	return result, nil
}

func (t *transaction) QueryRow(ctx context.Context, query string, args ...interface{}) Row {
	return t.Tx.QueryRowContext(ctx, query, args...)
}
