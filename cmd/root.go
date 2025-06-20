package cmd

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	c "github.com/seoyhaein/tori/config"
	dbUtils "github.com/seoyhaein/tori/db"
	globallog "github.com/seoyhaein/tori/log"
	"github.com/seoyhaein/tori/service"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var (
	cfg      = c.GlobalConfig
	logger   = globallog.Log
	database *sql.DB
)

// Execute 는 Cobra 루트 커맨드를 실행합니다.
func Execute() error {
	root := &cobra.Command{
		Use:   "tori-admin",
		Short: "관리자용 CLI for Tori service",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			database, err = dbUtils.ConnectDB("sqlite3", "file_monitor.db", true)
			if err != nil {
				return fmt.Errorf("DB 연결 실패: %w", err)
			}
			if err := dbUtils.InitializeDatabase(database); err != nil {
				return fmt.Errorf("DB 초기화 실패: %w", err)
			}
			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if database != nil {
				if err := database.Close(); err != nil {
					logger.Warnf("DB close 실패: %v", err)
				}
			}
		},
	}

	// 서브커맨드 등록
	root.AddCommand(
		serveCmd(),
		dumpCmd(),
		resetCmd(),
		snapshotCmd(),
		syncCmd(),
	)

	return root.Execute()
}

// TODO 일단 추후 구현.
func serveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "gRPC 서버 실행",
		RunE: func(cmd *cobra.Command, args []string) error {
			/*ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()
			logger.Info("gRPC 서버 시작 :50053")
			return api.ServeGRPC(ctx, database)*/
			return nil
		},
	}
}

func dumpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dump [output-file]",
		Short: "데이터블록을 텍스트 포맷으로 파일 저장",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			in := filepath.Join(cfg.RootDir, "datablock.pb")
			out := args[0]
			db, err := service.LoadDataBlock(in)
			if err != nil {
				return fmt.Errorf("데이터블록 로드 실패: %w", err)
			}
			if err := service.SaveDataBlockToTextFile(out, db); err != nil {
				return fmt.Errorf("텍스트 저장 실패: %w", err)
			}
			logger.Infof("Saved DataBlock to %s", out)
			return nil
		},
	}
}

func resetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset-db [db-file]",
		Short: "DB 파일 제거",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := os.Remove(args[0]); err != nil {
				return fmt.Errorf("DB 파일 삭제 실패: %w", err)
			}
			logger.Infof("Removed DB file %s", args[0])
			return nil
		},
	}
}

// snapshotCmd 는 현재 디렉터리 구조를 DB에 스냅샷으로 저장합니다.
func snapshotCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "snapshot",
		Short: "폴더 구조를 DB에 스냅샷 저장",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := dbUtils.SaveFolders(cmd.Context(), database, cfg.RootDir, cfg.FoldersExclusions, cfg.FilesExclusions); err != nil {
				return fmt.Errorf("스냅샷 저장 실패: %w", err)
			}
			logger.Info("폴더 구조 스냅샷 저장 완료")
			return nil
		},
	}
}

// syncCmd 는 DB 스냅샷과 실제 폴더를 비교·동기화합니다.
func syncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "스냅샷과 실제 폴더 비교 및 동기화",
		RunE: func(cmd *cobra.Command, args []string) error {
			updated, err := service.SyncFolders(cmd.Context())
			if err != nil {
				return fmt.Errorf("동기화 실패: %w", err)
			}
			if updated {
				logger.Info("폴더 변경 사항 반영 및 DataBlock 재생성 완료")
			} else {
				logger.Info("변경 사항 없음 – 동기화 생략")
			}
			return nil
		},
	}
}
