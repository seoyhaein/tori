package api

import (
	"context"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
	pb "github.com/seoyhaein/api-protos/gen/go/datablock/ichthys"
	"github.com/seoyhaein/api-protos/gen/go/datablock/ichthys/service"
	"github.com/seoyhaein/tori/config"
	d "github.com/seoyhaein/tori/db"
	"google.golang.org/protobuf/types/known/timestamppb"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// helper to create rule.json and sample files
func setupRuleDir(t *testing.T) (string, []string) {
	t.Helper()
	dir := t.TempDir()
	rs := map[string]any{
		"version":     "1",
		"delimiter":   []string{"_", ".txt"},
		"header":      []string{"H1", "H2"},
		"rowRules":    map[string]any{"matchParts": []int{0}},
		"columnRules": map[string]any{"matchParts": []int{1}},
		"sizeRules":   map[string]any{"minSize": 0, "maxSize": 1000},
	}
	b, _ := json.Marshal(rs)
	if err := os.WriteFile(filepath.Join(dir, "rule.json"), b, 0644); err != nil {
		t.Fatalf("write rule.json: %v", err)
	}
	files := []string{"r1_c1.txt", "r1_c2.txt"}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("x"), 0644); err != nil {
			t.Fatalf("write file: %v", err)
		}
	}
	return dir, files
}

func TestFileExistsExact(t *testing.T) {
	dir := t.TempDir()
	name := "test.txt"
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("data"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	exists, err := FileExistsExact(dir, name)
	if err != nil {
		t.Fatalf("FileExistsExact error: %v", err)
	}
	if !exists {
		t.Errorf("expected file to exist")
	}
}

func TestSearchFilesByPattern(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte(""), 0644)

	files, err := SearchFilesByPattern(dir, "*.txt")
	if err != nil {
		t.Fatalf("SearchFilesByPattern error: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}
}

func TestDeleteFilesByPattern(t *testing.T) {
	dir := t.TempDir()
	f1 := filepath.Join(dir, "a.txt")
	f2 := filepath.Join(dir, "b.txt")
	os.WriteFile(f1, []byte(""), 0644)
	os.WriteFile(f2, []byte(""), 0644)

	if err := DeleteFilesByPattern(dir, "*.txt"); err != nil {
		t.Fatalf("DeleteFilesByPattern error: %v", err)
	}
	if _, err := os.Stat(f1); !os.IsNotExist(err) {
		t.Errorf("expected %s to be removed", f1)
	}
	if _, err := os.Stat(f2); !os.IsNotExist(err) {
		t.Errorf("expected %s to be removed", f2)
	}
}

func TestDeleteFiles(t *testing.T) {
	dir := t.TempDir()
	f1 := filepath.Join(dir, "a.txt")
	f2 := filepath.Join(dir, "b.txt")
	if err := os.WriteFile(f1, []byte("a"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := os.WriteFile(f2, []byte("b"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := DeleteFiles([]string{f1, f2}); err != nil {
		t.Fatalf("DeleteFiles error: %v", err)
	}
	if _, err := os.Stat(f1); !os.IsNotExist(err) {
		t.Errorf("expected %s to be deleted", f1)
	}
	if _, err := os.Stat(f2); !os.IsNotExist(err) {
		t.Errorf("expected %s to be deleted", f2)
	}
}

func TestDeleteFiles_Single(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(f, []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := DeleteFiles([]string{f}); err != nil {
		t.Fatalf("DeleteFiles error: %v", err)
	}
	if _, err := os.Stat(f); err != nil {
		t.Errorf("file should not be deleted: %v", err)
	}
}

func TestSaveDataBlockToTextFile(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "db.txt")
	db := &pb.DataBlock{UpdatedAt: timestamppb.Now()}
	if err := SaveDataBlockToTextFile(out, db); err != nil {
		t.Fatalf("SaveDataBlockToTextFile error: %v", err)
	}
	info, err := os.Stat(out)
	if err != nil {
		t.Fatalf("output not created: %v", err)
	}
	if info.Size() == 0 {
		t.Errorf("expected file to be non-empty")
	}
}

func TestSGenerateDataBlock(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.pb")
	fb := &pb.FileBlock{
		BlockId:       "id1",
		ColumnHeaders: []string{"h"},
		Rows:          []*pb.Row{{RowNumber: 1, Cells: map[string]string{"h": "v"}}},
	}
	if err := GenerateDataBlock([]*pb.FileBlock{fb}, out); err != nil {
		t.Fatalf("SaveDataBlock error: %v", err)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("output file missing: %v", err)
	}
	dbLoaded, err := service.LoadDataBlock(out)
	if err != nil {
		t.Fatalf("failed to load datablock: %v", err)
	}
	if len(dbLoaded.Blocks) != 1 || dbLoaded.Blocks[0].BlockId != "id1" {
		t.Errorf("unexpected datablock contents")
	}
}

func TestGenerateFileBlock(t *testing.T) {
	dir, files := setupRuleDir(t)
	fb, err := GenerateFileBlock(dir, files)
	if err != nil {
		t.Fatalf("GenerateFileBlock error: %v", err)
	}
	if fb.BlockId != dir {
		t.Errorf("block id mismatch: %s", fb.BlockId)
	}
	if len(fb.Rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(fb.Rows))
	}
}

/*func TestConvertFolderFilesToFileBlocks(t *testing.T) {
	dir, files := setupRuleDir(t)
	ff := [][]string{{dir}}
	ff[0] = append(ff[0], files...)
	fbs, err := ConvertFolderFilesToFileBlocks(ff, []string{"H1", "H2"})
	if err != nil {
		t.Fatalf("ConvertFolderFilesToFileBlocks error: %v", err)
	}
	if len(fbs) != 1 {
		t.Fatalf("expected 1 fileblock, got %d", len(fbs))
	}
	if fbs[0].BlockId != dir {
		t.Errorf("block id mismatch")
	}
}*/

//  1. 폴더 내용과 DB가 동일할 때: SyncFolders는 변경 없음으로 간주하고,
//     최초 호출 시에도 DataBlock을 생성함을 확인합니다.
func TestSyncFolders_IdenticalFoldersAndDB(t *testing.T) {
	// 임시 디렉터리 생성 → RootDir로 설정
	tmpDir := t.TempDir()
	origRoot := config.GlobalConfig.RootDir
	config.GlobalConfig.RootDir = tmpDir
	defer func() { config.GlobalConfig.RootDir = origRoot }()

	// tmpDir 내부에 initial_folder/a.txt 생성
	initialFolder := filepath.Join(tmpDir, "initial_folder")
	if err := os.Mkdir(initialFolder, 0755); err != nil {
		t.Fatalf("초기 폴더 생성 실패: %v", err)
	}
	initialFile := filepath.Join(initialFolder, "a.txt")
	if err := os.WriteFile(initialFile, []byte("hello initial"), 0644); err != nil {
		t.Fatalf("초기 파일 생성 실패: %v", err)
	}

	// DB 파일 경로: tmpDir/file_monitor.db
	dbPath := filepath.Join(tmpDir, "file_monitor.db")

	// DB 연결 및 초기화
	dbConn, dbErr := d.ConnectDB("sqlite3", dbPath, true)
	if dbErr != nil {
		t.Fatalf("DB 연결 실패: %v", dbErr)
	}
	defer dbConn.Close()
	if dbErr = d.InitializeDatabase(dbConn); dbErr != nil {
		t.Fatalf("DB 초기화 실패: %v", dbErr)
	}

	// 서비스 인스턴스 생성
	f := &DataBlockServiceServerImpl{
		db:  dbConn,
		cfg: config.GlobalConfig,
	}
	ctx := context.Background()

	// SaveFolders 호출: tmpDir 내부 정보를 DB에 삽입
	if err := f.SaveFolders(ctx); err != nil {
		t.Fatalf("SaveFolders 호출 중 오류 발생: %v", err)
	}

	// 첫 번째 SyncFolders 호출: “DB와 FS(현재 tmpDir)가 동일” → 최초 DataBlock 생성
	got, err := f.SyncFolders(ctx)
	if err != nil {
		t.Fatalf("첫 번째 SyncFolders 호출 중 오류 발생: %v", err)
	}
	if !got {
		t.Fatalf("첫 번째 SyncFolders 반환값이 false; true여야 합니다")
	}

	// datablock.pb가 생성되었는지 확인
	pbFile := filepath.Join(tmpDir, "datablock.pb")
	info1, err := os.Stat(pbFile)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("첫 번째 SyncFolders 후 datablock.pb 파일이 생성되지 않았습니다")
		}
		t.Fatalf("datablock.pb 상태 확인 중 오류: %v", err)
	}
	if info1.Size() == 0 {
		t.Fatalf("생성된 datablock.pb 파일 크기가 0바이트입니다; 비어 있으면 안 됩니다")
	}

	// 두 번째 SyncFolders 호출 전 아무 변경 없이 재실행
	time.Sleep(10 * time.Millisecond) // ModTime 차이를 주기 위해 잠시 대기
	got2, err := f.SyncFolders(ctx)
	if err != nil {
		t.Fatalf("두 번째 SyncFolders 호출 중 오류 발생: %v", err)
	}
	if !got2 {
		t.Fatalf("두 번째 SyncFolders 반환값이 false; true여야 합니다")
	}

	// 두 번째 호출에서는 “DB와 FS가 여전히 동일”이므로 DataBlock 생성 시각이 동일하거나
	// 업데이트하지 않을 수 있음. 따라서 ModTime가 바뀌지 않아도 무방합니다.
	info2, err := os.Stat(pbFile)
	if err != nil {
		t.Fatalf("두 번째 호출 후 datablock.pb 상태 확인 실패: %v", err)
	}
	if !info2.ModTime().Equal(info1.ModTime()) && !info2.ModTime().After(info1.ModTime()) {
		t.Fatalf(
			"두 번째 호출 후 datablock.pb 업데이트 시각이 예상 범위를 벗어났습니다 (이전: %v, 이후: %v)",
			info1.ModTime(), info2.ModTime(),
		)
	}
}

// 2) 폴더 내용과 DB가 다를 때: SyncFolders는 차이를 감지하여 DataBlock을 갱신함을 확인합니다.
func TestSyncFolders_FoldersAndDBDiffer(t *testing.T) {
	// 임시 디렉터리 생성 → RootDir로 설정
	tmpDir := t.TempDir()
	origRoot := config.GlobalConfig.RootDir
	config.GlobalConfig.RootDir = tmpDir
	defer func() { config.GlobalConfig.RootDir = origRoot }()

	// tmpDir 내부에 initial_folder/a.txt 생성
	initialFolder := filepath.Join(tmpDir, "initial_folder")
	if err := os.Mkdir(initialFolder, 0755); err != nil {
		t.Fatalf("초기 폴더 생성 실패: %v", err)
	}
	initialFile := filepath.Join(initialFolder, "a.txt")
	if err := os.WriteFile(initialFile, []byte("hello initial"), 0644); err != nil {
		t.Fatalf("초기 파일 생성 실패: %v", err)
	}

	// DB 파일 경로: tmpDir/file_monitor.db
	dbPath := filepath.Join(tmpDir, "file_monitor.db")

	// DB 연결 및 초기화
	dbConn, dbErr := d.ConnectDB("sqlite3", dbPath, true)
	if dbErr != nil {
		t.Fatalf("DB 연결 실패: %v", dbErr)
	}
	defer dbConn.Close()
	if dbErr = d.InitializeDatabase(dbConn); dbErr != nil {
		t.Fatalf("DB 초기화 실패: %v", dbErr)
	}

	// 서비스 인스턴스 생성
	f := &DataBlockServiceServerImpl{
		db:  dbConn,
		cfg: config.GlobalConfig,
	}
	ctx := context.Background()

	// SaveFolders 호출: tmpDir 내부 정보를 DB에 삽입
	if err := f.SaveFolders(ctx); err != nil {
		t.Fatalf("SaveFolders 호출 중 오류 발생: %v", err)
	}

	// 첫 번째 SyncFolders 호출: 최초 DataBlock 생성
	got, err := f.SyncFolders(ctx)
	if err != nil {
		t.Fatalf("첫 번째 SyncFolders 호출 중 오류 발생: %v", err)
	}
	if !got {
		t.Fatalf("첫 번째 SyncFolders 반환값이 false; true여야 합니다")
	}

	// datablock.pb가 생성되었는지 확인
	pbFile := filepath.Join(tmpDir, "datablock.pb")
	info1, err := os.Stat(pbFile)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("첫 번째 SyncFolders 후 datablock.pb 파일이 생성되지 않았습니다")
		}
		t.Fatalf("datablock.pb 상태 확인 중 오류: %v", err)
	}
	if info1.Size() == 0 {
		t.Fatalf("생성된 datablock.pb 파일 크기가 0바이트입니다; 비어 있으면 안 됩니다")
	}

	// 약간 대기하여 ModTime 차이를 명확히 함
	time.Sleep(10 * time.Millisecond)

	// tmpDir 내부에 새로운 폴더/파일 생성하여 “DB와 FS가 다르다”시나리오 만듦
	newFolder := filepath.Join(tmpDir, "new_folder")
	if err := os.Mkdir(newFolder, 0755); err != nil {
		t.Fatalf("새 폴더 생성 실패: %v", err)
	}
	newFile := filepath.Join(newFolder, "example.txt")
	if err := os.WriteFile(newFile, []byte("hello world"), 0644); err != nil {
		t.Fatalf("새 파일 생성 실패: %v", err)
	}

	// 두 번째 SyncFolders 호출: 변경 있음으로 감지되어 DataBlock 갱신
	got2, err := f.SyncFolders(ctx)
	if err != nil {
		t.Fatalf("두 번째 SyncFolders 호출 중 오류 발생: %v", err)
	}
	if !got2 {
		t.Fatalf("두 번째 SyncFolders 반환값이 false; true여야 합니다")
	}

	// datablock.pb 수정 시각이 업데이트되었는지 확인
	info2, err := os.Stat(pbFile)
	if err != nil {
		t.Fatalf("두 번째 호출 후 datablock.pb 상태 확인 실패: %v", err)
	}
	if !info2.ModTime().After(info1.ModTime()) {
		t.Fatalf(
			"두 번째 SyncFolders 호출 후에도 datablock.pb가 업데이트되지 않았습니다 (이전: %v, 이후: %v)",
			info1.ModTime(), info2.ModTime(),
		)
	}
}

// MakeTestFilesA는 지정한 폴더(path)에 직접 파일을 생성
func MakeTestFilesA(path string) {
	// 디렉터리 생성 (재귀)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		logger.Fatalf("Failed to create directory %s: %v", path, err)
	}

	// 예시 FASTQ 파일 이름
	fileNames := []string{
		"SRA_S1_L001_R1_001.fastq.gz",
		"SRA_S1_L001_R2_001.fastq.gz",
		"SRA_S1_L002_R1_001.fastq.gz",
		"SRA_S1_L002_R2_001.fastq.gz",
	}

	// 파일 생성
	for _, fileName := range fileNames {
		filePath := filepath.Join(path, fileName)
		f, err := os.Create(filePath)
		if err != nil {
			logger.Fatalf("Failed to create file %s: %v", filePath, err)
		}
		f.Close()
		logger.Infof("Created file: %s", filePath)
	}
}

// WriteRuleJSON은 지정한 디렉터리에 rule.json을 생성
func WriteRuleJSON(dir string) {
	rule := `{
  "version": "1.0.1",
  "delimiter": ["_", "."],
  "header": ["R1", "R2"],
  "rowRules": {
    "matchParts": [0, 1, 2, 4, 5, 6]
  },
  "columnRules": {
    "matchParts": [3]
  },
  "sizeRules": {
    "minSize": 100,
    "maxSize": 1048576
  }
}`
	rulePath := filepath.Join(dir, "rule.json")
	if err := os.WriteFile(rulePath, []byte(rule), 0644); err != nil {
		logger.Fatalf("Failed to write rule.json at %s: %v", rulePath, err)
	}
	logger.Infof("Created rule.json at: %s", rulePath)
}

func TestSyncFolders_WithRulesAndFiles(t *testing.T) {
	// 1) 테스트용 임시 디렉터리 생성 → 이것이 “루트 폴더”
	tmpDir := "/test01" // 일단 그냥 만들어 줬음.
	origRoot := config.GlobalConfig.RootDir
	config.GlobalConfig.RootDir = tmpDir
	defer func() { config.GlobalConfig.RootDir = origRoot }()

	// 1a) tmpDir 바로 아래에 “test_folder” 폴더를 만들고,
	//     그 안에 테스트 파일들과 rule.json을 생성
	testFolder := filepath.Join(tmpDir, "test_folder")
	MakeTestFilesA(testFolder) // tmpDir/test_folder에 FASTQ 파일들 생성
	WriteRuleJSON(testFolder)  // tmpDir/test_folder/rule.json 생성

	// 2) tmpDir 바로 아래에 datablock.pb가 생성될 것이므로,
	//    한번도 SyncFolders 호출 전에는 없다.
	//    DB 파일 경로는 tmpDir/file_monitor.db
	dbPath := filepath.Join(tmpDir, "file_monitor.db")

	// 3) DB 연결 및 초기화
	dbConn, dbErr := d.ConnectDB("sqlite3", dbPath, true)
	if dbErr != nil {
		t.Fatalf("DB 연결 실패: %v", dbErr)
	}
	defer dbConn.Close()
	if dbErr = d.InitializeDatabase(dbConn); dbErr != nil {
		t.Fatalf("DB 초기화 실패: %v", dbErr)
	}

	// 4) 서비스 인스턴스 생성 (cfg.RootDir == tmpDir)
	f := &DataBlockServiceServerImpl{
		db:  dbConn,
		cfg: config.GlobalConfig,
	}
	ctx := context.Background()

	// --- 초기 SaveFolders 호출: 루트(tmpDir) 밑에 있는 “test_folder” 내부 정보를 DB에 삽입 ---
	if err := f.SaveFolders(ctx); err != nil {
		t.Fatalf("SaveFolders 호출 중 오류 발생: %v", err)
	}

	// --- 첫 번째 SyncFolders 호출: 최초 실행이므로 datablock.pb를 생성해야 함 ---
	got, err := f.SyncFolders(ctx)
	if err != nil {
		t.Fatalf("첫 번째 SyncFolders 호출 중 오류 발생: %v", err)
	}
	if !got {
		t.Fatalf("첫 번째 SyncFolders 반환값이 false; true여야 합니다")
	}

	// tmpDir/datablock.pb가 생성되었는지 확인
	pbFile := filepath.Join(tmpDir, "datablock.pb")
	info1, err := os.Stat(pbFile)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("첫 번째 SyncFolders 후 datablock.pb 파일이 생성되지 않았습니다")
		}
		t.Fatalf("datablock.pb 상태 확인 중 오류: %v", err)
	}
	if info1.Size() == 0 {
		t.Fatalf("생성된 datablock.pb 파일 크기가 0바이트입니다; 비어 있으면 안 됩니다")
	}

	// 약간 대기하여 ModTime 차이를 명확히 함
	time.Sleep(10 * time.Millisecond)

	// --- 두 번째 호출 전: test_folder 내부에 새 폴더/파일 생성하여 “DB와 FS가 다름” 유발 ---
	newFolder := filepath.Join(tmpDir, "new_subfolder")

	if err := os.Mkdir(newFolder, 0755); err != nil {
		t.Fatalf("새 폴더 생성 실패: %v", err)
	}
	WriteRuleJSON(newFolder)
	newFile := filepath.Join(newFolder, "new_sample.fastq.gz")
	if err := os.WriteFile(newFile, []byte("hello world"), 0644); err != nil {
		t.Fatalf("새 파일 생성 실패: %v", err)
	}

	// --- 두 번째 SyncFolders 호출: 변경 있음으로 감지되어 datablock.pb 갱신 ---
	got2, err := f.SyncFolders(ctx)
	if err != nil {
		t.Fatalf("두 번째 SyncFolders 호출 중 오류 발생: %v", err)
	}
	if !got2 {
		t.Fatalf("두 번째 SyncFolders 반환값이 false; true여야 합니다")
	}

	// tmpDir/datablock.pb의 수정 시각이 업데이트되었는지 확인
	info2, err := os.Stat(pbFile)
	if err != nil {
		t.Fatalf("두 번째 호출 후 datablock.pb 상태 확인 실패: %v", err)
	}
	if !info2.ModTime().After(info1.ModTime()) {
		t.Fatalf(
			"두 번째 SyncFolders 호출 후에도 datablock.pb가 업데이트되지 않았습니다 (이전: %v, 이후: %v)",
			info1.ModTime(), info2.ModTime(),
		)
	}
}
