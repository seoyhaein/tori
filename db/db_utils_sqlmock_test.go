package db

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/DATA-DOG/go-sqlmock"
)

// 참고: "Unable to resolve table 'test'" 또는 다른 경고는 SQL 구문에서 참조하는 테이블이 실제 데이터베이스에 없어서 발생함.
// 이 테스트에서는 fstest.MapFS와 sqlmock을 사용하여 SQL 실행을 시뮬레이션하므로 실제 테이블 존재 여부는 무시해도 됨.
// testFS는 테스트용 가상 파일 시스템
var testFS = fstest.MapFS{
	"queries/test_valid.sql":        &fstest.MapFile{Data: []byte("INSERT INTO test (id) VALUES (1);")},
	"queries/test_empty.sql":        &fstest.MapFile{Data: []byte("")},
	"queries/test_fail.sql":         &fstest.MapFile{Data: []byte("UPDATE test SET id = 1;")},
	"queries/test_select_valid.sql": &fstest.MapFile{Data: []byte("SELECT 1;")},
	"queries/test_select_fail.sql":  &fstest.MapFile{Data: []byte("SELECT * FROM non_existing_table;")},
}

// 각 테스트 시작 전에 sqlFiles를 테스트용 파일 시스템으로 재정의합니다.
func initTestFS() {
	sqlFiles = testFS
}

// -------------------
// Transaction 관련 테스트 (execSQLTx, execSQLTxNoCtx)
// -------------------

// TestExecSQLTx_FileNotFound TODO 이 기준으로 나머지 테스트 메서드도 작성 해야함.
func TestExecSQLTx_FileNotFound(t *testing.T) {
	initTestFS()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer func() {
		cErr := db.Close()
		if cErr != nil {
			Log.Warnf("failed to db close: %v", cErr)
		}
	}()

	// 트랜잭션 시작 전에 rollback 을 항상 호출하도록 defer 등록
	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	defer func(tx *sql.Tx) {
		tErr := tx.Rollback()
		if tErr != nil {
			Log.Warnf("failed to tx Rollback: %v", tErr)
		}
	}(tx)

	// 존재하지 않는 파일을 지정하여 에러가 발생하는지 검증
	err = execSQLTx(context.Background(), tx, "nonexistent.sql")
	if err == nil {
		t.Fatalf("expected error for nonexistent file, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read SQL file") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TODO 여기서 부터 시작.

func TestExecSQLTx_EmptyFile(t *testing.T) {
	initTestFS()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	err = execSQLTx(context.Background(), tx, "test_empty.sql")
	if err == nil {
		t.Errorf("expected error for empty SQL file, got nil")
	}
	if !strings.Contains(err.Error(), "is empty") {
		t.Errorf("unexpected error message: %v", err)
	}

	_ = tx.Rollback()
}

func TestExecSQLTx_Success(t *testing.T) {
	initTestFS()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	// 테스트용 SQL 파일을 읽고, 쿼리 문자열을 생성
	content, err := fs.ReadFile(sqlFiles, "queries/test_valid.sql")
	if err != nil {
		t.Fatalf("failed to read test_valid.sql: %v", err)
	}
	query := strings.TrimSpace(string(content))

	mock.ExpectExec(query).WillReturnResult(sqlmock.NewResult(1, 1))

	err = execSQLTx(context.Background(), tx, "test_valid.sql")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}

	_ = tx.Rollback()
}

func TestExecSQLTx_QueryExecutionError(t *testing.T) {
	initTestFS()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	content, err := fs.ReadFile(sqlFiles, "queries/test_fail.sql")
	if err != nil {
		t.Fatalf("failed to read test_fail.sql: %v", err)
	}
	query := strings.TrimSpace(string(content))

	expectedErr := errors.New("execution error")
	mock.ExpectExec(query).WillReturnError(expectedErr)

	err = execSQLTx(context.Background(), tx, "test_fail.sql")
	if err == nil {
		t.Errorf("expected error from ExecContext, got nil")
	}
	if !strings.Contains(err.Error(), "SQL execution failed") {
		t.Errorf("unexpected error message: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}

	_ = tx.Rollback()
}

func TestExecSQLTxNoCtx_Success(t *testing.T) {
	initTestFS()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	content, err := fs.ReadFile(sqlFiles, "queries/test_valid.sql")
	if err != nil {
		t.Fatalf("failed to read test_valid.sql: %v", err)
	}
	query := strings.TrimSpace(string(content))

	mock.ExpectExec(query).WillReturnResult(sqlmock.NewResult(1, 1))

	err = execSQLTxNoCtx(tx, "test_valid.sql")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}

	_ = tx.Rollback()
}

// -------------------
// DB 관련 테스트 (execSQL, execSQLNoCtx)
// -------------------

func TestExecSQL_FileNotFound(t *testing.T) {
	initTestFS()
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	err = execSQL(context.Background(), db, "nonexistent.sql")
	if err == nil {
		t.Errorf("expected error for nonexistent file, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read SQL file") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestExecSQL_EmptyFile(t *testing.T) {
	initTestFS()
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	err = execSQL(context.Background(), db, "test_empty.sql")
	if err == nil {
		t.Errorf("expected error for empty SQL file, got nil")
	}
	if !strings.Contains(err.Error(), "is empty") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestExecSQL_Success(t *testing.T) {
	initTestFS()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	content, err := fs.ReadFile(sqlFiles, "queries/test_valid.sql")
	if err != nil {
		t.Fatalf("failed to read test_valid.sql: %v", err)
	}
	query := strings.TrimSpace(string(content))

	mock.ExpectExec(query).WillReturnResult(sqlmock.NewResult(1, 1))

	err = execSQL(context.Background(), db, "test_valid.sql")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestExecSQL_QueryExecutionError(t *testing.T) {
	initTestFS()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	content, err := fs.ReadFile(sqlFiles, "queries/test_fail.sql")
	if err != nil {
		t.Fatalf("failed to read test_fail.sql: %v", err)
	}
	query := strings.TrimSpace(string(content))

	expectedErr := errors.New("execution error")
	mock.ExpectExec(query).WillReturnError(expectedErr)

	err = execSQL(context.Background(), db, "test_fail.sql")
	if err == nil {
		t.Errorf("expected error from ExecContext, got nil")
	}
	if !strings.Contains(err.Error(), "SQL execution failed") {
		t.Errorf("unexpected error message: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestExecSQLNoCtx_Success(t *testing.T) {
	initTestFS()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	content, err := fs.ReadFile(sqlFiles, "queries/test_valid.sql")
	if err != nil {
		t.Fatalf("failed to read test_valid.sql: %v", err)
	}
	query := strings.TrimSpace(string(content))

	mock.ExpectExec(query).WillReturnResult(sqlmock.NewResult(1, 1))

	err = execSQLNoCtx(db, "test_valid.sql")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

// -------------------
// SELECT 쿼리 관련 테스트 (querySQL, querySQLNoCtx)
// -------------------

func TestQuerySQL_FileNotFound(t *testing.T) {
	initTestFS()
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	_, err = querySQL(context.Background(), db, "nonexistent.sql")
	if err == nil {
		t.Errorf("expected error for nonexistent file, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read SQL file") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestQuerySQL_EmptyFile(t *testing.T) {
	initTestFS()
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	_, err = querySQL(context.Background(), db, "test_empty.sql")
	if err == nil {
		t.Errorf("expected error for empty SQL file, got nil")
	}
	if !strings.Contains(err.Error(), "is empty") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestQuerySQL_Success(t *testing.T) {
	initTestFS()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	content, err := fs.ReadFile(sqlFiles, "queries/test_select_valid.sql")
	if err != nil {
		t.Fatalf("failed to read test_select_valid.sql: %v", err)
	}
	query := strings.TrimSpace(string(content))

	// 모의 결과 행 생성
	rows := sqlmock.NewRows([]string{"col"}).AddRow(1)
	mock.ExpectQuery(query).WillReturnRows(rows)

	result, err := querySQL(context.Background(), db, "test_select_valid.sql")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	result.Close()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestQuerySQL_QueryError(t *testing.T) {
	initTestFS()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	content, err := fs.ReadFile(sqlFiles, "queries/test_select_fail.sql")
	if err != nil {
		t.Fatalf("failed to read test_select_fail.sql: %v", err)
	}
	query := strings.TrimSpace(string(content))

	expectedErr := errors.New("query error")
	mock.ExpectQuery(query).WillReturnError(expectedErr)

	_, err = querySQL(context.Background(), db, "test_select_fail.sql")
	if err == nil {
		t.Errorf("expected error from QueryContext, got nil")
	}
	if !strings.Contains(err.Error(), "SQL query failed") {
		t.Errorf("unexpected error message: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestQuerySQLNoCtx_Success(t *testing.T) {
	initTestFS()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	content, err := fs.ReadFile(sqlFiles, "queries/test_select_valid.sql")
	if err != nil {
		t.Fatalf("failed to read test_select_valid.sql: %v", err)
	}
	query := strings.TrimSpace(string(content))

	rows := sqlmock.NewRows([]string{"col"}).AddRow(1)
	mock.ExpectQuery(query).WillReturnRows(rows)

	result, err := querySQLNoCtx(db, "test_select_valid.sql")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	result.Close()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}
