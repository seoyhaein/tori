package db

import (
	"database/sql"
	"fmt"
)

// ConnectDB 데이터베이스에 연결하고, enableForeignKeys 가 true 이면 SQLite 사용 시 외래 키 제약 조건을 활성화함.
func ConnectDB(driverName, dataSourceName string, enableForeignKeys bool) (*sql.DB, error) {
	// SQLite 외의 드라이버는 지원하지 않음.
	if driverName != "sqlite3" {
		return nil, fmt.Errorf("unsupported driver: %s; only sqlite3 is supported", driverName)
	}

	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	// enableForeignKeys 가 true 면, SQLite 의 외래 키 제약 조건 활성화
	if enableForeignKeys {
		if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
			// 외래 키 활성화 실패 시 db.Close() 호출하고, 에러도 함께 처리함.
			if cErr := db.Close(); cErr != nil {
				return nil, fmt.Errorf("failed to enable foreign keys: %w; additionally failed to close db: %w", err, cErr)
			}
			return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
		}
	}
	return db, nil
}

// InitializeDatabase embed 된  SQL 파일(init.sql)을 사용하여 데이터베이스를 초기화
func InitializeDatabase(db *sql.DB) error {
	if !isDBInitialized(db) {
		logger.Info("Running DB initialization (embed)...")
		if err := execSQLNoCtx(db, "init.sql"); err != nil {
			return fmt.Errorf("DB initialization failed: %w", err)
		}
		logger.Info("DB initialization completed successfully (embed).")
	} else {
		logger.Info("DB already initialized. Skipping init.sql execution.")
	}
	return nil
}

func isDBInitialized(db *sql.DB) bool {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('folders', 'files')").Scan(&count)
	if err != nil {
		logger.Warnf("Failed to check database initialization:%v", err)
		return false
	}
	return count == 2 // folders, files 두 테이블이 모두 있어야 true 반환
}
