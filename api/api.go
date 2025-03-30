package api

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	c "github.com/seoyhaein/tori/config"
	d "github.com/seoyhaein/tori/db"
	globallog "github.com/seoyhaein/tori/log"
	pb "github.com/seoyhaein/tori/protos"
	"github.com/seoyhaein/tori/v1rpc"
	u "github.com/seoyhaein/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
	"os"
	"path/filepath"
)

// TODO 전역 변수 다른 곳엣 쓰는 문제 한번 생각해보자. 지금 코드가 중복되는 경향이 있음.

var (
	Config = c.GlobalConfig
	Db     = d.GlobalDb
	Log    = globallog.Log
)

// DBApis 데이터베이스와 관련된 인터페이스
type DBApis interface {
	StoreFoldersInfo(ctx context.Context, db *sql.DB) error
	CompareFoldersAndFiles(ctx context.Context, db *sql.DB) (*bool, []d.FolderDiff, []d.FileChange, []*pb.FileBlock, error)
	GetDataBlock(ctx context.Context, updateAt *timestamppb.Timestamp) (*pb.DataBlock, error)
}

// dBApisImpl DBApis 인터페이스의 구현체
type dBApisImpl struct {
	rootDir          string // 내부적으로 사용할 폴더 경로 (예: config.RootDir)
	foldersExclusion []string
	filesExclusions  []string
}

// NewDBApis DBApis 인터페이스의 구현체를 생성하는 factory 함수
func NewDBApis() DBApis {
	return &dBApisImpl{
		rootDir:          Config.RootDir,
		foldersExclusion: nil,
		filesExclusions:  Config.Exclusions,
	}
}

// StoreFoldersInfo 폴더 정보를 DB에 저장
func (f *dBApisImpl) StoreFoldersInfo(ctx context.Context, db *sql.DB) error {
	err := d.StoreFoldersInfo(ctx, db, f.rootDir, f.foldersExclusion, f.filesExclusions)
	return err
}

// CompareFoldersAndFiles 폴더와 파일을 비교하고, 변경 내역을 반환 TODO 수정해야 함 버그 있음. 분리한 메서드가 안정적일 경우 삭제 보관.
func (f *dBApisImpl) CompareFoldersAndFiles(ctx context.Context, db *sql.DB) (*bool, []d.FolderDiff, []d.FileChange, []*pb.FileBlock, error) {
	// 1. 폴더 비교: 폴더 목록과 폴더 간 차이 정보를 가져옴
	_, folders, folderDiffs, err := d.CompareFolders(db, f.rootDir, f.foldersExclusion, f.filesExclusions)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	var allFileChanges []d.FileChange
	var fbs []*pb.FileBlock // 파일 블록 데이터 슬라이스

	// 2. 각 폴더에 대해 파일 비교
	for _, folder := range folders {
		// 파일 비교
		filesMatch, files, fileChanges, err := d.CompareFiles(db, folder.Path, f.filesExclusions)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		// 해당 폴더에 파일이 다르다면 변경 내역에 추가
		if !filesMatch {
			allFileChanges = append(allFileChanges, fileChanges...)
		} else {
			// 파일과 폴더가 db와 동일한 경우, 특수 파일 존재 여부를 확인

			// rule.json 파일이 없으면 에러 리턴
			ruleExists, err := FileExistsExact(folder.Path, "rule.json")
			if !ruleExists {
				return nil, nil, nil, nil, fmt.Errorf("required file rule.json does not exist in folder %s", folder.Path)
			}
			if err != nil {
				return nil, nil, nil, nil, err
			}

			// fileblock.csv 존재 여부 확인
			bfb, err := FileExistsExact(folder.Path, "fileblock.csv")
			if err != nil {
				return nil, nil, nil, nil, err
			}

			// *.pb 존재 여부 확인
			pbs, err := SearchFilesByPattern(folder.Path, "*.pb")
			if err != nil {
				return nil, nil, nil, nil, err
			}

			// 만약 pb 파일이 여러 개이면 삭제 후 빈 슬라이스로 초기화
			if len(pbs) > 1 {
				if err = DeleteFiles(pbs); err != nil {
					return nil, nil, nil, nil, err
				}
				pbs = []string{}
			}

			// rule.json 있고, fileblock.csv 있으며, 정확히 하나의 pb 파일이 있으면 기존 파일 블록 로드
			if bfb && len(pbs) == 1 {
				pbPath := pbs[0]
				fb, err := v1rpc.LoadFileBlock(pbPath)
				if err != nil {
					return nil, nil, nil, nil, err
				}
				fbs = append(fbs, fb)
				continue
			}

			// []Files 를 []string(파일 이름 목록)으로 변환 후, 새 파일 블록 생성
			fileNames := d.ExtractFileNames(files)
			fb, err := v1rpc.GenerateFileBlock(folder.Path, fileNames)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			fbs = append(fbs, fb)
		}
	}

	// 전체 동일 여부 결정: 폴더 차이와 파일 변경 내역이 없으면 true, 아니면 false
	overallSame := len(folderDiffs) == 0 && len(allFileChanges) == 0
	if overallSame {
		return u.PTrue, nil, nil, fbs, nil
	}
	return u.PFalse, folderDiffs, allFileChanges, fbs, nil
}

// TODO 이렇게 분리 해놓았는데, 테스트를 좀 진행해야 할듯. 여기서 부터 시작 해야함. 중요.

func (f *dBApisImpl) CompareFoldersFiles(db *sql.DB) (*bool, [][]string, []d.FolderDiff, []d.FileChange, error) {
	// 1. 폴더 비교: 폴더 목록과 폴더 간 차이 정보를 가져옴
	_, folders, folderDiffs, err := d.CompareFolders(db, f.rootDir, f.foldersExclusion, f.filesExclusions)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	var (
		folderFiles    [][]string
		allFileChanges []d.FileChange
	)

	// 2. 각 폴더에 대해 파일 비교
	for _, folder := range folders {
		// 파일 비교
		filesMatch, files, fileChanges, err := d.CompareFiles(db, folder.Path, f.filesExclusions)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		// 해당 폴더에 파일이 다르다면 변경 내역에 추가
		if !filesMatch {
			allFileChanges = append(allFileChanges, fileChanges...)
		}

		// 폴더의 파일 목록을 추출하여 folderFiles 저장
		fileNames := d.ExtractFileNames(files)
		// folder.Path 를 key(첫번째 요소)로, 나머지 파일 이름들을 값으로 저장
		folderFiles = append(folderFiles, append([]string{folder.Path}, fileNames...))
	}

	// 전체 동일 여부 결정: 폴더 차이와 파일 변경 내역이 없으면 true, 아니면 false
	overallSame := len(folderDiffs) == 0 && len(allFileChanges) == 0
	if overallSame {
		return u.PTrue, folderFiles, nil, nil, nil
	}
	return u.PFalse, folderFiles, folderDiffs, allFileChanges, nil
}

// ConvertFolderFilesToFileBlocks converts a slice of folder-files ([][]string)
// into a slice of *pb.FileBlock. Each inner slice should have the first element
// as the folder path and the subsequent elements as file names.
// The provided headers will be assigned to the FileBlock.ColumnHeaders.
func ConvertFolderFilesToFileBlocks(folderFiles [][]string, headers []string) ([]*pb.FileBlock, error) {
	var fileBlocks []*pb.FileBlock

	for _, ff := range folderFiles {
		// ff가 비어있다면 건너뜀.
		if len(ff) == 0 {
			continue
		}
		// 첫번째 요소를 폴더 경로로 사용
		folderPath := ff[0]
		var fileNames []string
		if len(ff) > 1 {
			fileNames = ff[1:]
		}

		// 기존에 정의한 GenerateFileBlock 함수를 호출하여 FileBlock 생성
		fb, err := v1rpc.GenerateFileBlock(folderPath, fileNames)
		if err != nil {
			return nil, fmt.Errorf("failed to generate file block for folder %s: %w", folderPath, err)
		}
		// 전달받은 헤더를 할당
		fb.ColumnHeaders = headers
		fileBlocks = append(fileBlocks, fb)
	}
	return fileBlocks, nil
}

// UpdateFilesAndFolders 폴더 변경 내역과 파일 변경 내역을 DB에 반영
func UpdateFilesAndFolders(ctx context.Context, db *sql.DB, diffs []d.FolderDiff, changes []d.FileChange) error {
	// 폴더 변경 업데이트
	if err := d.UpsertFolders(ctx, db, diffs); err != nil {
		return err
	}
	// 파일 변경 업데이트
	if err := d.UpsertDelFiles(ctx, db, changes); err != nil {
		return err
	}
	return nil
}

// SaveDataBlock fileblock 을 병합하여 datablcok 으로 저장
// outputFile 은 파일이어야 함. 파일이 존재할 경우는 체크 하지 않고 덮어씀.
func SaveDataBlock(inputBlocks []*pb.FileBlock, outputFile string) error {
	dataBlock, err := v1rpc.MergeFileBlocksFromData(inputBlocks)
	if err != nil {
		return err
	}

	// DataBlock 저장
	if err := v1rpc.SaveProtoToFile(outputFile, dataBlock, os.ModePerm); err != nil {
		return fmt.Errorf("failed to save DataBlock: %w", err)
	}

	fmt.Printf("Successfully merged %d FileBlock files into %s\n", len(inputBlocks), outputFile)
	return nil
}

// FileExistsExact 주어진 폴더 내에서 정확한 파일명이 존재하는지 확인. 별도로 FileExists 가 있지만 그냥 이걸 씀.
func FileExistsExact(folder, fileName string) (bool, error) {
	path := filepath.Join(folder, fileName)
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, fmt.Errorf("파일 체크 실패 (%s): %w", path, err)
	}
}

// SearchFilesByPattern 주어진 폴더 내에서 지정한 glob 패턴에 매칭되는 파일들을 검색
// 검색 결과로 매칭된 파일 경로들의 배열을 반환합니다.
func SearchFilesByPattern(folder, pattern string) ([]string, error) {
	fullPattern := filepath.Join(folder, pattern)
	matches, err := filepath.Glob(fullPattern)
	if err != nil {
		return nil, fmt.Errorf("패턴 검색 실패 (%s): %w", pattern, err)
	}
	return matches, nil
}

// DeleteFilesByPattern 주어진 폴더 내에서 지정한 glob 패턴에 매칭되는 파일들을 검색해서 삭제함
// 만약 매칭된 파일이 2개 이상이면, 해당 파일들을 모두 삭제
func DeleteFilesByPattern(folder, pattern string) error {
	files, err := SearchFilesByPattern(folder, pattern)
	if err != nil {
		return fmt.Errorf("패턴 검색 실패 (%s): %w", pattern, err)
	}

	// 매칭된 파일이 여러 개인 경우에만 삭제 수행
	if len(files) > 1 {
		for _, filePath := range files {
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("파일 삭제 실패 (%s): %w", filePath, err)
			}
		}
	}
	return nil
}

// DeleteFiles 전달받은 파일 경로 목록에서 2개 이상의 파일이 존재하면 모두 삭제
func DeleteFiles(files []string) error {
	if len(files) > 1 {
		for _, filePath := range files {
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("failed to delete file (%s): %w", filePath, err)
			}
		}
	}
	return nil
}

func (f *dBApisImpl) GetDataBlock(ctx context.Context, updateAt *timestamppb.Timestamp) (*pb.DataBlock, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	// 서버의 데이터 블록 경로 정리
	dataBlockPath := filepath.Clean(f.rootDir)
	dataBlockPath = filepath.Join(dataBlockPath, "datablock.pb")

	// 서버의 데이터 블록 로드
	dataBlock, err := v1rpc.LoadDataBlock(dataBlockPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load datablock from %s: %w", dataBlockPath, err)
	}

	// 서버 데이터 블록에 UpdatedAt 필드가 없는 경우 에러 처리
	if dataBlock.UpdatedAt == nil {
		return nil, fmt.Errorf("server datablock is missing UpdatedAt field")
	}

	// 클라이언트가 업데이트 타임스탬프를 제공하지 않으면, 서버 데이터를 반환
	if updateAt == nil {
		return dataBlock, nil
	}

	// 클라이언트와 서버 타임스탬프를 Go의 time.Time 으로 변환
	clientTime := updateAt.AsTime()
	serverTime := dataBlock.UpdatedAt.AsTime()

	if clientTime.Before(serverTime) {
		// 클라이언트 데이터가 구버전이면 서버의 최신 데이터를 반환
		return dataBlock, nil
	} else if clientTime.Equal(serverTime) {
		// 버전이 동일하면 업데이트할 내용이 없으므로 nil 반환
		return nil, nil
	} else { // clientTime.After(serverTime)
		// 클라이언트 데이터가 서버보다 최신하면 에러 반환
		return nil, fmt.Errorf("client datablock is newer than server version")
	}
}

// SyncFoldersInfo  업데이트가 이루어졌으면 true, 그렇지 않으면 false TODO 여기 수정 중.
func SyncFoldersInfo(ctx context.Context, force bool) (bool, error) {

	dbApis := NewDBApis()
	err := dbApis.StoreFoldersInfo(ctx, Db)
	if err != nil {
		return false, fmt.Errorf("failed to store folders info into db : %v", err)
	}
	// force 가 true 이면 동기화 시킴.
	if force {

	}

	b, fDiff, fChange, fb, err := dbApis.CompareFoldersAndFiles(ctx, Db)
	if err != nil {
		Log.Fatalf("폴더와 파일을 비교하고, 변경 내역을 반환 실패: %v", err)
	}

	if b != nil && *b {
		// 전체 폴더와 파일이 동일한 경우 (b가 true)
		fmt.Println("모든 폴더와 파일이 동일합니다.")
		// 여기서 fileBlocks 등 추가 처리를 할 수 있습니다.
	} else if b != nil && !*b {
		if err = UpdateFilesAndFolders(ctx, Db, fDiff, fChange); err != nil {
			os.Exit(1)
		}
	} else {
		// err 가 not nil 이면 b 는 nil 임. 중복됨.
		os.Exit(1)
	}
	// fileblock 을 merge 해서 datablcok 으로 만들고 이후 파일로 저장함.
	outputDatablock := filepath.Join(Config.RootDir, "datablock.pb")
	if err = SaveDataBlock(fb, outputDatablock); err != nil {
		os.Exit(1)
	}

	return false, nil
}
