package main

import (
	"context"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	c "github.com/seoyhaein/tori/config"
	d "github.com/seoyhaein/tori/db"
	"os"
	"path"
	"path/filepath"
)

func main() {

	// 테스트 용도
	//_ = RemoveDBFile("file_monitor.db")
	// db connection foreign key 설정을 위해 PRAGMA foreign_keys = ON; 설정을 해줘야 함.
	db, err := d.ConnectDB("sqlite3", "file_monitor.db", true)
	if err != nil {
		os.Exit(1)
	}
	err = d.InitializeDatabase(db)
	if err != nil {
		os.Exit(1)
	}
	defer func() {
		/*if err = d.ClearDatabase(db); err != nil {
			//log.Fatal("failed to clear db:", err)
			os.Exit(1)
		}*/

		if err := db.Close(); err != nil {
			//log.Fatal("failed to close db:", err)
			os.Exit(1) // defer 내부에서도 os.Exit 사용 가능
		}
	}()

	config, err := c.LoadConfig("config.json")
	if err != nil {
		os.Exit(1)
	}

	// 테스트로 빈파일 생성
	// 기존 파일이 생성되어 있을 경우 권한 설정을 안해줌. 버그지만 고치지 않음.
	testFilePath := config.RootDir
	testFilePath = filepath.Join(testFilePath, "testFiles/")
	testFilePath = path.Clean(testFilePath)
	d.MakeTestFiles(testFilePath)
	d.MakeTestFilesA("/test/baba/")

	ctx := context.Background()
	// exclusion 은 보안상 여기다가 넣어둠. TODO 일단 생각은 해보자.
	exclusions := []string{"*.json", "invalid_files", "*.csv", "*.pb"}
	dbApis := NewDBApis(config.RootDir, nil, exclusions)
	err = dbApis.StoreFoldersInfo(ctx, db)
	if err != nil {

		os.Exit(1)
	}

	b, fDiff, fChange, fb, err := dbApis.CompareFoldersAndFiles(ctx, db)
	if err != nil {
		os.Exit(1)
	}

	if b != nil && *b {
		// 전체 폴더와 파일이 동일한 경우 (b가 true)
		fmt.Println("모든 폴더와 파일이 동일합니다.")
		// 여기서 fileBlocks 등 추가 처리를 할 수 있습니다.
	} else if b != nil && !*b {
		if err = UpdateFilesAndFolders(ctx, db, fDiff, fChange); err != nil {
			os.Exit(1)
		}
	} else {
		// err 가 not nil 이면 b 는 nil 임. 중복됨.
		os.Exit(1)
	}
	// fileblock 을 merge 해서 datablcok 으로 만들고 이후 파일로 저장함.
	outputDatablock := filepath.Join(config.RootDir, "datablock.pb")
	if err = SaveDataBlock(fb, outputDatablock); err != nil {
		os.Exit(1)
	}

}

// RemoveDBFile 주어진 DB 파일을 삭제함.
// filePath: 삭제할 DB 파일의 경로.
func RemoveDBFile(filePath string) error {
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to remove DB file %s: %w", filePath, err)
	}
	return nil
}
