package api

import (
	"context"
	"database/sql"
	"fmt"
	pb "github.com/seoyhaein/api-protos/gen/go/datablock/ichthys"
	"github.com/seoyhaein/api-protos/gen/go/datablock/ichthys/service"
	"github.com/seoyhaein/tori/config"
	dbUtils "github.com/seoyhaein/tori/db"
	globallog "github.com/seoyhaein/tori/log"
	"github.com/seoyhaein/tori/rules"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/types/known/timestamppb"
	"os"
	"path/filepath"
)

var logger = globallog.Log

type DataBlockServiceServerImpl struct {
	cfg *config.Config
	db  *sql.DB
}

// TODO 이름이 너무 김.

func NewDataBlockServiceServerImpl(db *sql.DB, cfg *config.Config) *DataBlockServiceServerImpl {
	return &DataBlockServiceServerImpl{
		db: db, cfg: cfg}
}

func (f *DataBlockServiceServerImpl) GetDataBlock(ctx context.Context, updateAt *timestamppb.Timestamp) (*pb.DataBlock, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	// 서버의 데이터 블록 경로 정리
	dataBlockPath := filepath.Clean(f.cfg.RootDir)
	dataBlockPath = filepath.Join(dataBlockPath, "datablock.pb")

	// 서버의 데이터 블록 로드
	dataBlock, err := LoadDataBlock(dataBlockPath)
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

// 추가 해줌.

// GenerateFileBlockFromDir 디렉터리 경로를 받아서 FileBlock 객체를 생성하고, 바이너리 protobuf 파일로 저장
func GenerateFileBlockFromDir(dirPath string) (*pb.FileBlock, error) {
	// 1. 룰 로딩
	ruleSet, err := rules.LoadRuleSetFromFile(dirPath)
	if err != nil {
		return nil, fmt.Errorf("LoadRuleSetFromFile error: %w", err)
	}

	// 2. 룰 검증
	if !rules.IsValidRuleSet(ruleSet) {
		return nil, fmt.Errorf("rule set validation failed")
	}

	// 3. 디렉터리 내 파일 목록 읽기 (제외 패턴 지정)
	exclusions := []string{"rule.json", "invalid_files", "fileblock.csv", "*.pb"}
	fileNames, err := rules.ListFilesExclude(dirPath, exclusions)
	if err != nil {
		return nil, fmt.Errorf("ReadAllFileNames error: %w", err)
	}

	// 4. 룰 기준으로 map[int]map[string]string 생성
	resultMap, err := rules.GroupFiles(fileNames, ruleSet)
	if err != nil {
		return nil, fmt.Errorf("GroupFiles error: %w", err)
	}

	// 5. 유효/무효 행 분리
	validMap, invalidRows := rules.FilterGroups(resultMap, len(ruleSet.Header))

	// 6. validMap → CSV 파일로 저장
	if err := rules.ExportResultsCSV(validMap, ruleSet.Header, dirPath); err != nil {
		return nil, fmt.Errorf("SaveResultMapToCSV error: %w", err)
	}

	// 7. invalidRows → invalid 파일에 기록
	if err := rules.SaveInvalidFiles(invalidRows, dirPath); err != nil {
		return nil, fmt.Errorf("WriteInvalidFiles error: %w", err)
	}

	// 8. validMap + headers → FileBlock 객체 생성
	fb := service.ConvertMapToFileBlock(validMap, ruleSet.Header, dirPath)

	// 9. FileBlock → 바이너리 protobuf 파일로 저장
	outPath := filepath.Join(dirPath, filepath.Base(dirPath)+"files.pb")
	if err := service.SaveProtoToFile(outPath, fb, 0o777); err != nil {
		return nil, fmt.Errorf("SaveProtoToFile error: %w", err)
	}

	return fb, nil
}

// GenerateFileBlock 일단 이름 고침. filePath 는 rule.josn 이 있는 위치이자 fileblock.csv, invalid_files, *.pb 파일 등이 가 저장될 위치.
func GenerateFileBlock(filePath string, files []string) (*pb.FileBlock, error) {
	// Load the rule set
	ruleSet, err := rules.LoadRuleSetFromFile(filePath) // 이 메서드에서 filepath 의 검증을 해줌.
	if err != nil {
		return nil, fmt.Errorf("failed to load rule set: %w", err)
	}

	// Validate the rule set
	if !rules.IsValidRuleSet(ruleSet) {
		return nil, fmt.Errorf("rule set has conflicts or unused parts")
	}

	resultMap, err := rules.GroupFiles(files, ruleSet)
	if err != nil {
		return nil, fmt.Errorf("failed to blockify files: %w", err)
	}

	// Filter the result map into valid and invalid rows. 열의 갯수 기준으로 유효/무효 행을 분리
	validRows, invalidRows := rules.FilterGroups(resultMap, len(ruleSet.Header))

	// Save valid rows to a CSV file. 사용자에게 보여주기 위함.
	if err := rules.ExportResultsCSV(validRows, ruleSet.Header, filePath); err != nil {
		return nil, fmt.Errorf("failed to save result map to CSV: %w", err)
	}

	// Save invalid rows to a separate file
	if err := rules.SaveInvalidFiles(invalidRows, filePath); err != nil {
		return nil, fmt.Errorf("failed to write invalid files: %w", err)
	}

	// blockId 를 filePath 로 잡아둠.
	fbd := service.ConvertMapToFileBlock(validRows, ruleSet.Header, filePath)
	pbName := filepath.Join(filePath, fmt.Sprintf("%sfiles.pb", filepath.Base(filePath)))
	err = service.SaveProtoToFile(pbName, fbd, 0777)
	if err != nil {
		return nil, fmt.Errorf("failed to save proto to file: %w", err)
	}

	return fbd, nil
}

// SaveFolders 폴더 정보를 DB에 저장, TODO 이건 한번만 실행되어야 하는 메서드 임. 이름을 이러한 맥락을 고려해서 넣어 주어야 할듯, 아래 구현되어 있는 메서드들도 이름 수정해줘야 할듯.
func (f *DataBlockServiceServerImpl) SaveFolders(ctx context.Context) error {
	err := dbUtils.SaveFolders(ctx, f.db, f.cfg.RootDir, nil, f.cfg.Exclusions)
	return err
}

// SyncFolders 업데이트가 이루어졌으면 true, 그렇지 않으면 false
func (f *DataBlockServiceServerImpl) SyncFolders(ctx context.Context) (bool, error) {
	// 1) DiffFolders 호출
	folderFiles, fDiff, fChange, err := dbUtils.DiffFolders(f.db)
	if err != nil {
		logger.Errorf("DiffFolders 실패: %v", err)
		return false, err
	}

	// 2) 변경 없음 시 바로 반환 (FileBlock/DB 업데이트 생략)
	if fDiff == nil && fChange == nil {
		logger.Info("all files and folders are same; skipping update and DataBlock generation.")
		return true, nil
	}

	// 3) DB 업데이트
	if err = dbUtils.UpdateDB(ctx, f.db, fDiff, fChange); err != nil {
		logger.Errorf("UpdateDB 실패: %v", err)
		return false, err
	}
	// 컨텍스트 취소 여부 확인
	if ctx.Err() != nil {
		logger.Warnf("SyncFolders 종료: 컨텍스트 취소 감지 (%v)", ctx.Err())
		return false, ctx.Err()
	}

	// 4) FileBlock 생성
	fbs, err := GenerateFBs(folderFiles)
	if err != nil {
		logger.Errorf("GenerateFBs 실패: %v", err)
		return false, err
	}
	// 컨텍스트 취소 여부 재확인
	if ctx.Err() != nil {
		logger.Warnf("SyncFolders 종료: 컨텍스트 취소 감지 (%v)", ctx.Err())
		return false, ctx.Err()
	}

	// 5) DataBlock 저장
	outputDatablock := filepath.Join(f.cfg.RootDir, "datablock.pb")
	if err = GenerateDataBlock(fbs, outputDatablock); err != nil {
		logger.Errorf("GenerateDataBlock 실패 (%s): %v", outputDatablock, err)
		return false, err
	}
	// 최종 컨텍스트 취소 여부 확인
	if ctx.Err() != nil {
		logger.Warnf("SyncFolders 완료 이후 컨텍스트 취소 감지 (%v)", ctx.Err())
		return false, ctx.Err()
	}

	return true, nil
}

// GenerateFBs
func GenerateFBs(folderFiles [][]string) ([]*pb.FileBlock, error) {
	var fileBlocks []*pb.FileBlock

	for _, ff := range folderFiles {
		if len(ff) == 0 {
			continue
		}
		folderPath := ff[0]

		var fileNames []string
		if len(ff) > 1 {
			fileNames = ff[1:]
		}

		fb, err := GenerateFileBlock(folderPath, fileNames)
		if err != nil {
			return nil, fmt.Errorf("failed to generate file block for folder %s: %w", folderPath, err)
		}

		fileBlocks = append(fileBlocks, fb)
	}
	return fileBlocks, nil
}

// GenerateDataBlock fileblock 을 병합하여 datablcok 으로 저장
// outputFile 은 파일이어야 함. 파일이 존재할 경우는 체크 하지 않고 덮어씀.
func GenerateDataBlock(inputBlocks []*pb.FileBlock, outputFile string) error {
	dataBlock, err := service.MergeFileBlocksFromData(inputBlocks)
	if err != nil {
		return err
	}

	// DataBlock 저장
	if err := service.SaveProtoToFile(outputFile, dataBlock, os.ModePerm); err != nil {
		return fmt.Errorf("failed to save DataBlock: %w", err)
	}

	fmt.Printf("Successfully merged %d FileBlock files into %s\n", len(inputBlocks), outputFile)
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

// SaveDataBlockToTextFile DataBlockData 텍스트 포맷으로 파일에 저장
func SaveDataBlockToTextFile(filePath string, data *pb.DataBlock) error {
	// proto 메시지를 텍스트 포맷으로 변환
	textData, err := prototext.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal DataBlock to text format: %w", err)
	}

	// 텍스트 데이터를 파일에 저장
	if err := os.WriteFile(filePath, textData, os.ModePerm); err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filePath, err)
	}

	fmt.Printf("Successfully saved DataBlock to %s\n", filePath)
	return nil
}

func LoadDataBlock(filePath string) (*pb.DataBlock, error) {
	return service.LoadDataBlock(filePath)
}
