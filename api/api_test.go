package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	pb "github.com/seoyhaein/api-protos/gen/go/datablock/ichthys"
	"github.com/seoyhaein/api-protos/gen/go/datablock/ichthys/service"
	"google.golang.org/protobuf/types/known/timestamppb"
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

func TestSaveDataBlock(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.pb")
	fb := &pb.FileBlock{
		BlockId:       "id1",
		ColumnHeaders: []string{"h"},
		Rows:          []*pb.Row{{RowNumber: 1, Cells: map[string]string{"h": "v"}}},
	}
	if err := SaveDataBlock([]*pb.FileBlock{fb}, out); err != nil {
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

func TestConvertFolderFilesToFileBlocks(t *testing.T) {
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
}
