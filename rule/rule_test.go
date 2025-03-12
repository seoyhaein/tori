package rule_test

import (
	"fmt"
	"testing"

	"github.com/seoyhaein/tori/rule"
)

// 테스트를 위한 데이터 및 함수
func TestBlockifyFilesToMap(t *testing.T) {
	// 테스트 입력 데이터
	fileNames := []string{
		"sample1_S1_L001_R1_001.fastq.gz",
		"sample1_S1_L001_R2_001.fastq.gz",
		"sample1_S1_L002_R1_001.fastq.gz",
		"sample2_S2_L001_R1_001.fastq.gz",
		"sample2_S2_L001_R2_001.fastq.gz",
	}

	// JSON 규칙 설정
	ruleSet := rule.RuleSet{
		Delimiter: []string{"_", "."},
		Header:    []string{"R1", "R2"},
		RowRules: rule.RowRules{
			MatchParts: []int{0, 1, 2, 4}, // Row 기준 파트
		},
		ColumnRules: rule.ColumnRules{
			MatchParts: []int{3}, // Column 기준 파트
		},
	}

	// 기대 결과
	/*expected := map[int]map[string]string{
		0: {
			"R1": "sample1_S1_L001_R1_001.fastq.gz",
			"R2": "sample1_S1_L001_R2_001.fastq.gz",
		},
		1: {
			"R1": "sample1_S1_L002_R1_001.fastq.gz",
		},
		2: {
			"R1": "sample2_S2_L001_R1_001.fastq.gz",
			"R2": "sample2_S2_L001_R2_001.fastq.gz",
		},
	}*/

	// 함수 호출
	result, err := rule.BlockifyFilesToMap(fileNames, ruleSet)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// 결과 비교
	/*if !reflect.DeepEqual(result, expected) {
		t.Errorf("Result mismatch.\nExpected: %v\nGot: %v", expected, result)
	}*/

	err = rule.SaveResultMapToCSV("output.csv", result, ruleSet.Header)
	if err != nil {
		fmt.Println("Error saving CSV:", err)
	}
}

func TestBlockifyFilesToMap01(t *testing.T) {
	var ruleSet rule.RuleSet
	ruleSet.Version = "1.0.1"
	ruleSet.Delimiter = []string{"_", "."}
	ruleSet.Header = []string{"file1", "file2"}
	ruleSet.RowRules.MatchParts = []int{0, 1, 2, 4, 5, 6}
	ruleSet.ColumnRules.MatchParts = []int{3}
	ruleSet.SizeRules.MinSize = 100
	ruleSet.SizeRules.MaxSize = 1048576

	/*fileNames := []string{
		"sample1_S1_L001_R1_001.fastq.gz",
		"sample1_S1_L001_R2_001.fastq.gz",
	}*/

	incompleteFileNames := []string{
		"sample1_S1_L001_R1_001.fastq.gz",
		"sample1_S1_L001_R2_001.fastq.gz",
		"sample13_S13_L001_R1.fastq.gz",
		"sample14_S14_L001_R2_001.fastq",
		"sample15_S15_L001_001.fastq.gz",
		"sample16_S16_L001.fastq.gz",
	}

	//result, err := rule.BlockifyFilesToMap(fileNames, ruleSet)

	result, err := rule.BlockifyFilesToMap(incompleteFileNames, ruleSet)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	// 결과 검사 로직 (예: 특정 행/열에 기대한 파일명이 들어있는지)
	// 여기서는 단순히 로그로 확인
	for rowIdx, rowData := range result {
		t.Logf("Row %d:", rowIdx)
		for colKey, file := range rowData {
			t.Logf("  %s -> %s", colKey, file)
		}
	}

	// 실제 테스트에서는 특정 값을 기대하는지 검증하는 로직 추가 가능
}

/*func TestValidateRuleSet(t *testing.T) {
	tests := []struct {
		name          string
		ruleSet       rule.RuleSet
		fileName      string
		expectedError string
	}{
		{
			name: "Valid RuleSet with no conflicts and all parts covered",
			ruleSet: rule.RuleSet{
				Version:   "1.0.0",
				Delimiter: []string{"_", "."},
				Header:    []string{"R1", "R2"},
				RowRules: rule.RowRules{
					MatchParts:   []int{0, 1, 2, 4},
					UnMatchParts: []int{5},
				},
				ColumnRules: rule.ColumnRules{
					MatchParts:   []int{3},
					UnMatchParts: []int{6},
				},
			},
			fileName:      "sample1_S1_L001_R1_001.fastq.gz",
			expectedError: "",
		},
		{
			name: "Conflict between RowRules.MatchParts and ColumnRules.MatchParts",
			ruleSet: rule.RuleSet{
				Version:   "1.0.0",
				Delimiter: []string{"_", "."},
				Header:    []string{"R1", "R2"},
				RowRules: rule.RowRules{
					MatchParts: []int{0, 1, 2, 3}, // Part 3 conflicts
				},
				ColumnRules: rule.ColumnRules{
					MatchParts: []int{3}, // Conflict with RowRules
				},
			},
			fileName:      "sample1_S1_L001_R1_001.fastq.gz",
			expectedError: "conflict detected: part 3 is in both RowRules.MatchParts and ColumnRules.MatchParts",
		},
		{
			name: "Conflict between RowRules.UnMatchParts and ColumnRules.MatchParts",
			ruleSet: rule.RuleSet{
				Version:   "1.0.0",
				Delimiter: []string{"_", "."},
				Header:    []string{"R1", "R2"},
				RowRules: rule.RowRules{
					UnMatchParts: []int{3}, // Part 3 conflicts
				},
				ColumnRules: rule.ColumnRules{
					MatchParts: []int{3}, // Conflict with RowRules
				},
			},
			fileName:      "sample1_S1_L001_R1_001.fastq.gz",
			expectedError: "conflict detected: part 3 is in RowRules.UnMatchParts and ColumnRules.MatchParts",
		},
		{
			name: "Unused parts detected",
			ruleSet: rule.RuleSet{
				Version:   "1.0.0",
				Delimiter: []string{"_", "."},
				Header:    []string{"R1", "R2"},
				RowRules: rule.RowRules{
					MatchParts: []int{0, 1, 2, 4},
				},
				ColumnRules: rule.ColumnRules{
					MatchParts: []int{3},
				},
			},
			fileName:      "sample1_S1_L001_R1_001.fastq.gz",
			expectedError: "unused parts detected: [5 6]",
		},
		{
			name: "Empty MatchParts in RowRules and ColumnRules",
			ruleSet: rule.RuleSet{
				Version:   "1.0.0",
				Delimiter: []string{"_", "."},
				Header:    []string{"R1", "R2"},
				RowRules: rule.RowRules{
					MatchParts:   []int{},
					UnMatchParts: []int{},
				},
				ColumnRules: rule.ColumnRules{
					MatchParts:   []int{},
					UnMatchParts: []int{},
				},
			},
			fileName:      "sample1_S1_L001_R1_001.fastq.gz",
			expectedError: "unused parts detected: [0 1 2 3 4 5 6]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rule.ValidateRuleSet(tt.ruleSet)
			if !result {
				t.Errorf("Expected valid RuleSet but got invalid: %v", tt.ruleSet)
			}
		})
	}
}

func TestValidateRuleSet1(t *testing.T) {
	tests := []struct {
		name          string
		ruleSet       rule.RuleSet
		fileName      string
		expectedError string
	}{
		{
			name: "Conflict between RowRules.UnMatchParts and ColumnRules.MatchParts",
			ruleSet: rule.RuleSet{
				Version:   "1.0.0",
				Delimiter: []string{"_", "."},
				Header:    []string{"R1", "R2"},
				RowRules: rule.RowRules{
					UnMatchParts: []int{3}, // Part 3 conflicts
				},
				ColumnRules: rule.ColumnRules{
					MatchParts: []int{3}, // Conflict with RowRules
				},
			},
			fileName: "sample1_S1_L001_R1_001.fastq.gz",
			// 충돌과 Row/Col 규칙이 비어 있는지 함께 확인
			expectedError: "conflict detected: part 3 is in RowRules.UnMatchParts and ColumnRules.MatchParts; RowRules.MatchParts is empty",
		},
		{
			name: "No MatchParts or UnMatchParts in RowRules",
			ruleSet: rule.RuleSet{
				Version:   "1.0.0",
				Delimiter: []string{"_", "."},
				Header:    []string{"R1", "R2"},
				RowRules: rule.RowRules{
					MatchParts:   []int{}, // Empty MatchParts
					UnMatchParts: []int{}, // Empty UnMatchParts
				},
				ColumnRules: rule.ColumnRules{
					MatchParts:   []int{3},
					UnMatchParts: []int{},
				},
			},
			fileName:      "sample1_S1_L001_R1_001.fastq.gz",
			expectedError: "RowRules.MatchParts is empty; RowRules.UnMatchParts is empty",
		},
		{
			name: "No MatchParts or UnMatchParts in ColumnRules",
			ruleSet: rule.RuleSet{
				Version:   "1.0.0",
				Delimiter: []string{"_", "."},
				Header:    []string{"R1", "R2"},
				RowRules: rule.RowRules{
					MatchParts: []int{0, 1, 2},
				},
				ColumnRules: rule.ColumnRules{
					MatchParts:   []int{}, // Empty MatchParts
					UnMatchParts: []int{}, // Empty UnMatchParts
				},
			},
			fileName:      "sample1_S1_L001_R1_001.fastq.gz",
			expectedError: "ColumnRules.MatchParts is empty; ColumnRules.UnMatchParts is empty",
		},
		{
			name: "Unused parts detected",
			ruleSet: rule.RuleSet{
				Version:   "1.0.0",
				Delimiter: []string{"_", "."},
				Header:    []string{"R1", "R2"},
				RowRules: rule.RowRules{
					MatchParts: []int{0, 1, 2, 4},
				},
				ColumnRules: rule.ColumnRules{
					MatchParts: []int{3},
				},
			},
			fileName:      "sample1_S1_L001_R1_001.fastq.gz",
			expectedError: "unused parts detected: [5 6]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rule.ValidateRuleSet(tt.ruleSet)
			if !result {
				t.Errorf("Expected valid RuleSet but got invalid: %v", tt.ruleSet)
			}

		})
	}
}*/
