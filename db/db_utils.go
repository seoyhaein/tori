package db

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	globallog "github.com/seoyhaein/tori/log"
	u "github.com/seoyhaein/utils"
	"io/fs"
	"strings"
)

var (
	logger = globallog.Log
	//go:embed queries/*.sql
	embeddedFiles embed.FS
	// 중요 sqlFiles 를 fs.FS 타입으로 선언해서 테스트 시 fstest.MapFS 할당이 가능하도록 함.
	sqlFiles fs.FS = embeddedFiles
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

// db utils

// execSQLTx 읽어온 SQL 파일을 트랜잭션 내에서 ExecContext 로 실행.
// IMPORTANT: 비 SELECT 쿼리에 사용. (결과 리턴 없음) 호출하는 쪽에서 트랜젝션의 commit 이나 rollback 을 신경써줘야 함.
func execSQLTx(ctx context.Context, tx *sql.Tx, fileName string, args ...interface{}) error {
	// "queries/" 하위의 SQL 파일을 읽어옴.
	content, err := fs.ReadFile(sqlFiles, "queries/"+fileName)
	if err != nil {
		return fmt.Errorf("failed to read SQL file (%s): %w", fileName, err)
	}

	query := strings.TrimSpace(string(content))
	if query == "" {
		return fmt.Errorf("SQL file (%s) is empty", fileName)
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("SQL execution failed (%s): %w", fileName, err)
	}

	return nil
}

// execSQLTxNoCtx 컨텍스트 없이 트랜잭션 내에서 SQL 파일을 실행.
func execSQLTxNoCtx(tx *sql.Tx, fileName string, args ...interface{}) error {
	return execSQLTx(context.Background(), tx, fileName, args...)
}

// execSQL 읽어온 SQL 파일을 DB 에서 ExecContext 로 실행.
// IMPORTANT: 비 SELECT 쿼리에 사용. (결과 리턴 없음) 호출하는 쪽에서 트랜젝션의 commit 이나 rollback 을 신경써줘야 함.
func execSQL(ctx context.Context, db *sql.DB, fileName string, args ...interface{}) error {
	// "queries/" 하위의 SQL 파일을 읽어옴.
	content, err := fs.ReadFile(sqlFiles, "queries/"+fileName)
	if err != nil {
		return fmt.Errorf("failed to read SQL file (%s): %w", fileName, err)
	}

	query := strings.TrimSpace(string(content))
	if query == "" {
		return fmt.Errorf("SQL file (%s) is empty", fileName)
	}

	_, err = db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("SQL execution failed (%s): %w", fileName, err)
	}

	return nil
}

// execSQLNoCtx 컨텍스트 없이 DB 에서 SQL 파일을 실행.
func execSQLNoCtx(db *sql.DB, fileName string, args ...interface{}) error {
	return execSQL(context.Background(), db, fileName, args...)
}

// querySQL 읽어온 SQL 파일을 DB 에서 QueryContext 로 실행.
// IMPORTANT: SELECT 쿼리에 사용. 결과로 *sql.Rows 를 반환하며, 호출자가 반드시 Close() 해야 함.  않하면 memory leak 발생.
func querySQL(ctx context.Context, db *sql.DB, fileName string, args ...interface{}) (*sql.Rows, error) {
	// "queries/" 하위의 SQL 파일을 읽어옴.
	content, err := fs.ReadFile(sqlFiles, "queries/"+fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read SQL file (%s): %w", fileName, err)
	}

	query := strings.TrimSpace(string(content))
	if query == "" {
		return nil, fmt.Errorf("SQL file (%s) is empty", fileName)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("SQL query failed (%s): %w", fileName, err)
	}
	return rows, nil
}

// querySQLNoCtx 컨텍스트 없이 DB 에서 SELECT 쿼리를 실행.
// IMPORTANT: SELECT 쿼리에 사용. 결과로 *sql.Rows 를 반환하며, 호출자가 반드시 Close() 해야 함. 않하면 memory leak 발생.
func querySQLNoCtx(db *sql.DB, fileName string, args ...interface{}) (*sql.Rows, error) {
	return querySQL(context.Background(), db, fileName, args...)
}

// StoreFilesFolderInfo 폴더 경로를 받아 폴더 내 파일 정보를 DB에 삽입하는 함수, TODO 한번만 실행되고 말아야 함. 이름 수정하자.
func StoreFilesFolderInfo(ctx context.Context, db *sql.DB, folderPath string, exclusions []string) error {
	folderPath, err := u.CheckPath(folderPath)
	if err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	folderDetails, fileDetails, err := GetCurrentFolderFileInfo(folderPath, exclusions)

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
			logger.Infof("rollback failed: %v", rbErr)
		}
		return fmt.Errorf("failed to get folder details: %w", err)
	}

	// DB에 폴더 정보 삽입 (insert_folder.sql)
	err = execSQLTx(ctx, tx, "insert_folder.sql",
		folderDetails.Path,
		folderDetails.TotalSize,
		folderDetails.FileCount,
		folderDetails.CreatedTime)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
			logger.Infof("rollback failed: %v", rbErr)
		}
		return fmt.Errorf("failed to insert folder: %w", err)
	}

	// 삽입된 폴더의 ID를 조회
	var folderID int64
	err = tx.QueryRowContext(ctx, "SELECT id FROM folders WHERE path = ?", folderDetails.Path).Scan(&folderID)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
			logger.Infof("rollback failed: %v", rbErr)
		}
		return fmt.Errorf("failed to query folder ID: %w", err)
	}

	// 파일 정보 삽입 (insert_file.sql)
	for _, file := range fileDetails {
		err = execSQLTx(ctx, tx, "insert_file.sql",
			folderID,
			file.Name,
			file.Size,
			file.CreatedTime)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
				logger.Infof("rollback failed: %v", rbErr)
			}
			return fmt.Errorf("failed to insert file: %w", err)
		}
	}

	err = execSQLTx(ctx, tx, "update_folders_fromDB.sql", folderID)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
			logger.Infof("rollback failed: %v", rbErr)
		}
		return fmt.Errorf("failed to update folder statistics: %w", err)
	}

	// 트랜잭션 커밋
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func getFolderID(db *sql.DB, path string) (int64, error) {
	rows, err := querySQLNoCtx(db, "get_folder_id.sql", path)
	if err != nil {
		return 0, fmt.Errorf("querySQLNoCtx failed (get_folder_id.sql): %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			logger.Warnf("failed to close rows: %v", err)
		}
	}(rows)

	if !rows.Next() {
		return 0, sql.ErrNoRows
	}
	var id int64
	if err := rows.Scan(&id); err != nil {
		return 0, fmt.Errorf("failed to scan folder id: %w", err)
	}
	return id, nil
}

// UpsertFolders FolderDiff 슬라이스에 대해 DB 업데이트(업서트)를 수행.
func UpsertFolders(ctx context.Context, db *sql.DB, diffs []FolderDiff) error {
	for _, diff := range diffs {
		if err := diff.UpsertFolder(ctx, db); err != nil {
			return err
		}
	}
	return nil
}

// UpsertDelFiles 전에 []FileChange 에 folder_id 와 file id 를 채워 넣는 과정이 필요하다.

// UpsertDelFiles FileChange 슬라이스에 대해 DB 업데이트(업서트)를 수행.
func UpsertDelFiles(ctx context.Context, db *sql.DB, changes []FileChange) error {
	for _, change := range changes {
		if err := change.UpsertDelFile(ctx, db); err != nil {
			return err
		}
	}
	return nil
}

// CheckForeignKeysEnabled DB 연결에서 외래 키가 활성화되었는지 확인함.
// IMPORTANT: fk 값이 1이면 외래 키가 활성화된 상태임.
func CheckForeignKeysEnabled(db *sql.DB) (bool, error) {
	var fk int
	err := db.QueryRow("PRAGMA foreign_keys").Scan(&fk)
	if err != nil {
		return false, fmt.Errorf("failed to check foreign keys: %w", err)
	}
	return fk == 1, nil
}

// ClearDatabase for test
func ClearDatabase(db *sql.DB) error {
	// 외래 키 제약 조건이 ON DELETE CASCADE 로 설정되어 있다면, folders 테이블에서 데이터를 삭제하면 files 테이블의 데이터도 자동 삭제.
	_, err := db.Exec("DELETE FROM folders;")
	return err
}

// ExtractFileNames 변환.
func ExtractFileNames(files []File) []string {
	names := make([]string, 0, len(files))
	for _, f := range files {
		names = append(names, f.Name)
	}
	return names
}
