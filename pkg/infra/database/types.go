package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Row represents a database row
type Row interface {
	Scan(dest ...interface{}) error
}

// Result represents the result of a database operation
type Result struct {
	LastInsertId int64
	RowsAffected int64
}

// Transaction represents a database transaction
type Transaction interface {
	Exec(ctx context.Context, query string, args ...interface{}) (*Result, error)
	Query(ctx context.Context, query string, args ...interface{}) ([]Row, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) Row
	Commit() error
	Rollback() error
}

// Stats represents database statistics
type Stats struct {
	OpenConnections int
	InUse           int
	Idle            int
	WaitCount       int64
	WaitDuration    time.Duration
	MaxIdleTime     time.Duration
}

// Error represents a database error
type Error struct {
	Operation string
	Query     string
	Err       error
}

func (e *Error) Error() string {
	if e.Query != "" {
		return fmt.Sprintf("%s: %s: %v", e.Operation, e.Query, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Operation, e.Err)
}

// IsNoRows returns true if the error is sql.ErrNoRows
func IsNoRows(err error) bool {
	if err == sql.ErrNoRows {
		return true
	}
	if dbErr, ok := err.(*Error); ok {
		return dbErr.Err == sql.ErrNoRows
	}
	return false
}

// IsDuplicate returns true if the error is a duplicate key error
func IsDuplicate(err error) bool {
	if dbErr, ok := err.(*Error); ok {
		// Check MySQL error code 1062 (duplicate entry)
		return dbErr.Err.Error()[0:4] == "1062"
	}
	return false
}
