package db

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"
	"testing"
)

// TODO 내부 사용 api 들 테스트 진행해야함.

func TestIsDBInitialized(t *testing.T) {
	// in-memory SQLite 데이터베이스 생성
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("in-memory DB 생성 실패: %v", err)
	}
	defer func() {
		cErr := db.Close()
		if cErr != nil {
			Log.Warnf("failed to db close:%v", cErr)
		}
	}()

	// 초기 상태: 아무 테이블도 없으므로 false 여야 함.
	if isDBInitialized(db) {
		t.Error("테이블이 없는 DB 에서 isDBInitialized 가 true 를 반환함")
	}

	// folders 테이블만 생성
	_, err = db.Exec("CREATE TABLE folders (id INTEGER PRIMARY KEY)")
	if err != nil {
		t.Fatalf("folders 테이블 생성 실패: %v", err)
	}

	// files 테이블이 없으므로 여전히 false 여야 함.
	if isDBInitialized(db) {
		t.Error("folders 테이블만 존재할 때 isDBInitialized 가 true 를 반환함")
	}

	// files 테이블 생성
	_, err = db.Exec("CREATE TABLE files (id INTEGER PRIMARY KEY)")
	if err != nil {
		t.Fatalf("files 테이블 생성 실패: %v", err)
	}

	// 이제 folders, files 두 테이블이 모두 존재하므로 true 여야 함.
	if !isDBInitialized(db) {
		t.Error("folders 와 files 테이블이 모두 존재함에도 불구하고 isDBInitialized 가 false 를 반환함")
	}
}

func createTestFolder(root, folderName string) (string, error) {
	folderPath := filepath.Join(root, folderName)
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create folder %q: %w", folderPath, err)
	}
	filePath := filepath.Join(folderPath, "file.txt")
	content := []byte("test content")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return "", fmt.Errorf("failed to create file %q: %w", filePath, err)
	}
	return folderPath, nil
}

func insertFolderInfo(db *sql.DB, folder Folder) error {
	query := "INSERT INTO folders (path, total_size, file_count, created_time) VALUES (?, ?, ?, ?)"
	_, err := db.Exec(query, folder.Path, folder.TotalSize, folder.FileCount, folder.CreatedTime)
	return err
}

// TODO 일단 테스트 방법 부터 일단 고민해보자.
// InitializeDatabase 테스트 해야함.
// insertFolderInfo

func TestCompareFolders(t *testing.T) {
	tempRoot, err := os.MkdirTemp("", "test_compare_folders")
	if err != nil {
		t.Fatalf("failed to create temp root: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempRoot); err != nil {
			Log.Warnf("failed to RemoveAll: %v", err)
		}
	}()

	_, err = createTestFolder(tempRoot, "folder1")
	if err != nil {
		t.Fatalf("failed to create test folder: %v", err)
	}

	dbConn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory DB: %v", err)
	}
	defer func() {
		if err := dbConn.Close(); err != nil {
			Log.Warnf("failed to close DB: %v", err)
		}
	}()

	// 종속성인 InitializeDatabase 실패 시 테스트를 건너뜁니다.
	if err := InitializeDatabase(dbConn); err != nil {
		t.Skipf("Skipping tests because InitializeDatabase failed: %v", err)
	}

	var foldersExclusions, filesExclusions []string

	t.Run("DB empty - differences exist", func(t *testing.T) {
		unchanged, diskFolders, diffs, err := CompareFolders(dbConn, tempRoot, foldersExclusions, filesExclusions)
		if err != nil {
			t.Fatalf("CompareFolders returned error: %v", err)
		}

		if len(diskFolders) == 0 {
			t.Errorf("expected at least 1 disk folder, got 0")
		}

		if unchanged {
			t.Errorf("expected overallSame to be false, got true")
		}
		if len(diffs) == 0 {
			t.Errorf("expected folder differences, got none")
		}
	})

	t.Run("DB matches disk - no differences", func(t *testing.T) {
		diskFolders, err := GetFoldersInfo(tempRoot, foldersExclusions)
		if err != nil {
			t.Fatalf("GetFoldersInfo failed: %v", err)
		}
		if len(diskFolders) == 0 {
			t.Fatalf("no disk folders found")
		}

		err = insertFolderInfo(dbConn, diskFolders[0])
		if err != nil {
			t.Fatalf("failed to insert folder info into DB: %v", err)
		}

		unchanged, _, diffs, err := CompareFolders(dbConn, tempRoot, foldersExclusions, filesExclusions)
		if err != nil {
			t.Fatalf("CompareFolders returned error: %v", err)
		}

		if !unchanged {
			t.Errorf("expected overallSame to be true, got false")
		}
		if len(diffs) != 0 {
			t.Errorf("expected no folder differences, got: %v", diffs)
		}
	})
}
