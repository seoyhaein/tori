package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	c "github.com/seoyhaein/tori/config"
	d "github.com/seoyhaein/tori/db"
	globallog "github.com/seoyhaein/tori/log"
	"github.com/seoyhaein/tori/v1rpc"
	"github.com/seoyhaein/tori/v1rpc/server"
	"os"
	"path"
	"path/filepath"
)

var (
	db          *sql.DB
	serverErrCh chan error
	logger      = globallog.Log
	gConfig     = c.GlobalConfig
	btest       bool
)

func init() {
	// db 연결 및 설정
	var dbErr error
	db, dbErr = d.ConnectDB("sqlite3", "file_monitor.db", true)
	if dbErr != nil {
		logger.Fatalf("fail to connect sqlite3 %v", dbErr)
	}
	if dbErr = d.InitializeDatabase(db); dbErr != nil {
		logger.Fatalf("fail to initialize sqlite3 %v", dbErr)
	}

	//grpc 서버 시작.
	server.Address = ":50053"
	// gRPC 서버를 별도 고루틴에서 실행하고, 종료 에러를 받을 채널 생성
	serverErrCh = make(chan error, 1)
	go func() {
		err := server.Server()
		serverErrCh <- err
	}()

}

func main() {
	// Important 기억하자. defer 구문안에는 Exit 처리 하면 안됨.
	defer func() {
		if db != nil {
			if cErr := db.Close(); cErr != nil {
				//log.Fatal("failed to close db:", err)

				logger.Warnf("failed to db closed : %v ", cErr) // defer 내부에서도 os.Exit 사용 가능
			}
		}
	}()

	// TODO 확인하자.
	Shutdown()

	// 테스트로 빈파일 생성
	// 기존 파일이 생성되어 있을 경우 권한 설정을 안해줌. 버그지만 고치지 않음.
	// TODO 테스트 위해서 만들어둠. true 이면 생성해주고 false 이면 생성안해주는 것으로.
	btest = false
	if btest {
		testFilePath := gConfig.RootDir
		testFilePath = filepath.Join(testFilePath, "testFiles/")
		testFilePath = path.Clean(testFilePath)
		d.MakeTestFiles(testFilePath)
		d.MakeTestFilesA("/test/baba/")
	}

	ctx := context.Background()

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
		logger.Errorf("Server shutdown returned error: %v", err)
	}
}
