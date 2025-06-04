package rules

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSplitFileName(t *testing.T) {
	got := splitFileName("sample1_S1_L001_R1_001.fastq.gz", []string{"_", ".fastq.gz"})
	want := []string{"sample1", "S1", "L001", "R1", "001"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("splitFileName mismatch. got %v want %v", got, want)
	}
}

func TestGroupFilesAndFilterGroups(t *testing.T) {
	files := []string{
		"sample1_S1_L001_R1_001.fastq.gz",
		"sample1_S1_L002_R1_001.fastq.gz",
		"sample2_S2_L001_R1_001.fastq.gz",
	}
	rs := RuleSet{
		Delimiter:   []string{"_", ".fastq.gz"},
		RowRules:    RowRules{MatchParts: []int{0, 1}},
		ColumnRules: ColumnRules{MatchParts: []int{2}},
	}
	grouped, err := GroupFiles(files, rs)
	if err != nil {
		t.Fatalf("GroupFiles returned error: %v", err)
	}
	valid, invalid := FilterGroups(grouped, 2)
	if len(valid) != 1 {
		t.Errorf("expected 1 valid group, got %d", len(valid))
	}
	if len(invalid) != 1 {
		t.Errorf("expected 1 invalid group, got %d", len(invalid))
	}
}

func TestIsValidRuleSet(t *testing.T) {
	rs := RuleSet{
		RowRules:    RowRules{MatchParts: []int{0, 1}},
		ColumnRules: ColumnRules{MatchParts: []int{1}},
	}
	if IsValidRuleSet(rs) {
		t.Errorf("expected rule set to be invalid due to duplicate index")
	}
}

func TestListFilesExclude(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "keep.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, "skip.json"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, "invalid_files"), []byte(""), 0644)

	files, err := ListFilesExclude(dir, []string{"*.json", "invalid_files"})
	if err != nil {
		t.Fatalf("ListFilesExclude error: %v", err)
	}
	if len(files) != 1 || files[0] != "keep.txt" {
		t.Errorf("unexpected files: %v", files)
	}
}

func TestSaveInvalidFiles(t *testing.T) {
	dir := t.TempDir()
	rows := []map[string]string{
		{"a": "f1"},
		{"b": "f2"},
	}
	if err := SaveInvalidFiles(rows, dir); err != nil {
		t.Fatalf("SaveInvalidFiles error: %v", err)
	}
	matches, err := filepath.Glob(filepath.Join(dir, "invalid_files_*.txt"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("expected invalid file: %v %v", matches, err)
	}
	data, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("failed to read result: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}
}

func TestSaveInvalidFiles_NoRows(t *testing.T) {
	dir := t.TempDir()
	if err := SaveInvalidFiles(nil, dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	matches, _ := filepath.Glob(filepath.Join(dir, "invalid_files_*.txt"))
	if len(matches) != 0 {
		t.Errorf("expected no output file, got %d", len(matches))
	}
}

func TestExportResultsCSV(t *testing.T) {
	dir := t.TempDir()
	result := map[int]map[string]string{
		0: {"A": "a.txt", "B": "b.txt"},
		1: {"A": "c.txt", "B": "d.txt"},
	}
	headers := []string{"A", "B"}
	if err := ExportResultsCSV(result, headers, dir); err != nil {
		t.Fatalf("ExportResultsCSV error: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "fileblock.csv"))
	if err != nil {
		t.Fatalf("failed to read csv: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
	if !strings.Contains(lines[1], "a.txt") {
		t.Errorf("csv content unexpected: %v", lines[1])
	}
}

func TestLoadRuleSetFromFile(t *testing.T) {
	dir := t.TempDir()
	rs := RuleSet{Delimiter: []string{"_"}, Header: []string{"A"}, RowRules: RowRules{MatchParts: []int{0}}, ColumnRules: ColumnRules{MatchParts: []int{0}}}
	b, err := json.Marshal(rs)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "rule.json"), b, 0644); err != nil {
		t.Fatalf("write rule.json error: %v", err)
	}
	loaded, err := LoadRuleSetFromFile(dir)
	if err != nil {
		t.Fatalf("LoadRuleSetFromFile error: %v", err)
	}
	if loaded.Delimiter[0] != "_" || loaded.Header[0] != "A" {
		t.Errorf("loaded data mismatch: %+v", loaded)
	}
}
