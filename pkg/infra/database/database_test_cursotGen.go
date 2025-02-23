package database

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"order-system/pkg/infra/config"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.Config
		mock    func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "successful connection",
			cfg: &config.Config{
				Database: struct {
					Host         string        `json:"host"`
					Port         int           `json:"port"`
					User         string        `json:"user"`
					Password     string        `json:"password"`
					Database     string        `json:"database"`
					MaxOpenConns int           `json:"maxOpenConns"`
					MaxIdleConns int           `json:"maxIdleConns"`
					MaxLifetime  time.Duration `json:"maxLifetime"`
				}{
					Host:         "localhost",
					Port:         3306,
					User:         "test",
					Password:     "test",
					Database:     "test",
					MaxOpenConns: 10,
					MaxIdleConns: 5,
					MaxLifetime:  time.Hour,
				},
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing()
				mock.ExpectClose()
			},
			wantErr: false,
		},
		{
			name: "ping fails",
			cfg: &config.Config{
				Database: struct {
					Host         string        `json:"host"`
					Port         int           `json:"port"`
					User         string        `json:"user"`
					Password     string        `json:"password"`
					Database     string        `json:"database"`
					MaxOpenConns int           `json:"maxOpenConns"`
					MaxIdleConns int           `json:"maxIdleConns"`
					MaxLifetime  time.Duration `json:"maxLifetime"`
				}{
					Host:         "localhost",
					Port:         3306,
					User:         "test",
					Password:     "test",
					Database:     "test",
					MaxOpenConns: 10,
					MaxIdleConns: 5,
					MaxLifetime:  time.Hour,
				},
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing().WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
		{
			name: "invalid configuration",
			cfg: &config.Config{
				Database: struct {
					Host         string        `json:"host"`
					Port         int           `json:"port"`
					User         string        `json:"user"`
					Password     string        `json:"password"`
					Database     string        `json:"database"`
					MaxOpenConns int           `json:"maxOpenConns"`
					MaxIdleConns int           `json:"maxIdleConns"`
					MaxLifetime  time.Duration `json:"maxLifetime"`
				}{
					Host:         "",
					Port:         0,
					User:         "",
					Password:     "",
					Database:     "",
					MaxOpenConns: -1,
					MaxIdleConns: -1,
					MaxLifetime:  -1,
				},
			},
			mock: func(mock sqlmock.Sqlmock) {
				// 无效配置应该在 Ping 之前就返回错误
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 为每个测试用例创建新的 mock
			mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
			require.NoError(t, err)

			// 保存原始的 sql.Open 函数
			originalOpen := sqlOpen
			defer func() {
				sqlOpen = originalOpen
				mockDB.Close()
			}()

			// 替换 sql.Open 函数
			if tt.name == "invalid configuration" {
				sqlOpen = func(driverName, dataSourceName string) (*sql.DB, error) {
					return nil, sql.ErrConnDone
				}
			} else {
				sqlOpen = func(driverName, dataSourceName string) (*sql.DB, error) {
					return mockDB, nil
				}
			}

			// 设置 mock 预期
			tt.mock(mock)

			// 执行测试
			db, err := New(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, db)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, db)
				if db != nil {
					assert.NoError(t, db.Close())
				}
			}

			// 验证所有预期都被满足
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDB_Exec(t *testing.T) {
	// 创建 mock
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	// 创建测试用的 db 实例
	db := &db{
		DB:     mockDB,
		config: &config.Config{},
	}

	// 测试成功场景
	t.Run("successful exec", func(t *testing.T) {
		query := "INSERT INTO users (name) VALUES (?)"
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("test").
			WillReturnResult(sqlmock.NewResult(1, 1))

		result, err := db.Exec(context.Background(), query, "test")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		if result != nil {
			assert.Equal(t, int64(1), result.LastInsertId)
			assert.Equal(t, int64(1), result.RowsAffected)
		}
	})

	// 测试执行错误
	t.Run("exec error", func(t *testing.T) {
		query := "INSERT INTO users (name) VALUES (?)"
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("test").
			WillReturnError(sql.ErrConnDone)

		result, err := db.Exec(context.Background(), query, "test")
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	// 验证所有预期都被满足
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_Query(t *testing.T) {
	// 创建 mock
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	// 创建测试用的 db 实例
	db := &db{
		DB:     mockDB,
		config: &config.Config{},
	}

	// 测试成功场景
	t.Run("successful query", func(t *testing.T) {
		query := "SELECT id, name FROM users"
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "test1").
			AddRow(2, "test2")
		mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(rows)

		result, err := db.Query(context.Background(), query)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 2)
	})

	// 测试查询错误
	t.Run("query error", func(t *testing.T) {
		query := "SELECT id, name FROM users"
		mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(sql.ErrConnDone)

		result, err := db.Query(context.Background(), query)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	// 验证所有预期都被满足
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_QueryRow(t *testing.T) {
	// 创建 mock
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	// 创建测试用的 db 实例
	db := &db{
		DB:     mockDB,
		config: &config.Config{},
	}

	// 测试成功场景
	t.Run("successful query row", func(t *testing.T) {
		query := "SELECT id, name FROM users WHERE id = ?"
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "test")
		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(1).
			WillReturnRows(rows)

		row := db.QueryRow(context.Background(), query, 1)
		assert.NotNil(t, row)

		var id int
		var name string
		err := row.Scan(&id, &name)
		assert.NoError(t, err)
		assert.Equal(t, 1, id)
		assert.Equal(t, "test", name)
	})

	// 测试未找到数据
	t.Run("no rows", func(t *testing.T) {
		query := "SELECT id, name FROM users WHERE id = ?"
		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(999).
			WillReturnError(sql.ErrNoRows)

		row := db.QueryRow(context.Background(), query, 999)
		assert.NotNil(t, row)

		var id int
		var name string
		err := row.Scan(&id, &name)
		assert.Equal(t, sql.ErrNoRows, err)
	})

	// 验证所有预期都被满足
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_Transaction(t *testing.T) {
	// 创建 mock
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	// 创建测试用的 db 实例
	db := &db{
		DB:     mockDB,
		config: &config.Config{},
	}

	// 测试成功场景
	t.Run("successful transaction", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (name) VALUES (?)")).
			WithArgs("test").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := db.Transaction(context.Background(), func(tx Transaction) error {
			_, err := tx.Exec(context.Background(), "INSERT INTO users (name) VALUES (?)", "test")
			return err
		})
		assert.NoError(t, err)
	})

	// 测试回滚场景
	t.Run("rollback transaction", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (name) VALUES (?)")).
			WithArgs("test").
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		err := db.Transaction(context.Background(), func(tx Transaction) error {
			_, err := tx.Exec(context.Background(), "INSERT INTO users (name) VALUES (?)", "test")
			return err
		})
		assert.Error(t, err)
	})

	// 测试开始事务失败
	t.Run("begin transaction error", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(sql.ErrConnDone)

		err := db.Transaction(context.Background(), func(tx Transaction) error {
			return nil
		})
		assert.Error(t, err)
	})

	// 测试提交失败
	t.Run("commit error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (name) VALUES (?)")).
			WithArgs("test").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit().WillReturnError(sql.ErrConnDone)

		err := db.Transaction(context.Background(), func(tx Transaction) error {
			_, err := tx.Exec(context.Background(), "INSERT INTO users (name) VALUES (?)", "test")
			return err
		})
		assert.Error(t, err)
	})

	// 验证所有预期都被满足
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_Stats(t *testing.T) {
	// 创建 sqlmock
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	// 创建测试用的 db 实例
	db := &db{
		DB: mockDB,
		config: &config.Config{
			Database: struct {
				Host         string        `json:"host"`
				Port         int           `json:"port"`
				User         string        `json:"user"`
				Password     string        `json:"password"`
				Database     string        `json:"database"`
				MaxOpenConns int           `json:"maxOpenConns"`
				MaxIdleConns int           `json:"maxIdleConns"`
				MaxLifetime  time.Duration `json:"maxLifetime"`
			}{
				MaxOpenConns: 10,
				MaxIdleConns: 5,
				MaxLifetime:  time.Hour,
			},
		},
	}

	// 获取统计信息
	stats := db.Stats()

	// 验证统计信息结构是否完整
	assert.NotNil(t, stats)
	assert.IsType(t, Stats{}, stats)

	// 验证字段类型
	assert.IsType(t, 0, stats.OpenConnections)
	assert.IsType(t, 0, stats.InUse)
	assert.IsType(t, 0, stats.Idle)
	assert.IsType(t, int64(0), stats.WaitCount)
	assert.IsType(t, time.Duration(0), stats.WaitDuration)
	assert.IsType(t, time.Duration(0), stats.MaxIdleTime)
}
