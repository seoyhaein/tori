package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	c "github.com/seoyhaein/tori/config"
	d "github.com/seoyhaein/tori/db"
	"github.com/seoyhaein/tori/v1rpc"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"path/filepath"
)

var (
	db          *sql.DB
	serverErrCh chan error
	Log         = logrus.New()
)

func init() {
	// db 연결 및 설정
	var dbErr error
	db, dbErr = d.ConnectDB("sqlite3", "file_monitor.db", true)
	if dbErr != nil {
		Log.Fatalf("fail to connect sqlite3 %v", dbErr)
	}
	if dbErr = d.InitializeDatabase(db); dbErr != nil {
		Log.Fatalf("fail to initialize sqlite3 %v", dbErr)
	}

	//grpc 서버 시작.
	v1rpc.Address = ":50053"

	// gRPC 서버를 별도 고루틴에서 실행하고, 종료 에러를 받을 채널 생성
	serverErrCh = make(chan error, 1)
	go func() {
		err := v1rpc.Server()
		serverErrCh <- err
	}()
}

func main() {
	// Important 기억하자. defer 구문안에는 Exit 처리 하면 안됨.
	defer func() {
		if db != nil {
			if cErr := db.Close(); cErr != nil {
				//log.Fatal("failed to close db:", err)

				Log.Warnf("failed to db closed : %v ", cErr) // defer 내부에서도 os.Exit 사용 가능
			}
		}
	}()

	// TODO 확인하자.
	Shutdown()

	config, err := c.LoadConfig("config.json")
	// Important 기억하자. os.Exit(1) 로만 하지 말고 Log.Fatalf 를 써서 오류 사항을 명확히 하자. 자체적으로 os.Exit(1) 처리됨.
	if err != nil {
		Log.Fatalf("failed to load config file %v", err)
	}
	// 전역 config 설정.
	c.GlobalConfig = config

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

	dpb, _ := v1rpc.LoadDataBlock(outputDatablock)
	v1rpc.SaveDataBlockToTextFile("/test/datablock.txt", dpb)

}

// RemoveDBFile 주어진 DB 파일을 삭제함.
// filePath: 삭제할 DB 파일의 경로.
func RemoveDBFile(filePath string) error {
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to remove DB file %s: %w", filePath, err)
	}
	return nil
}

func Shutdown() {
	err := <-serverErrCh
	if err != nil {
		Log.Errorf("Server shutdown returned error: %v", err)
	}
}
