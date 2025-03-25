package v1rpc

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	pb "github.com/seoyhaein/tori/protos"
	u "github.com/seoyhaein/utils"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// RuleSet 구조체 정의
type RuleSet struct {
	Version     string      `json:"version"`
	Delimiter   []string    `json:"delimiter"`
	Header      []string    `json:"header"`
	RowRules    RowRules    `json:"rowRules"`
	ColumnRules ColumnRules `json:"columnRules"`
	SizeRules   SizeRules   `json:"sizeRules"`
}

type RowRules struct {
	MatchParts []int `json:"matchParts"`
}

type ColumnRules struct {
	MatchParts []int `json:"matchParts"`
}

type SizeRules struct {
	MinSize int `json:"minSize"`
	MaxSize int `json:"maxSize"`
}

// LoadRuleSetFromFile JSON 파일을 읽어 RuleSet 구조체로 디코딩. RuleSet 의 경우 값의 수정이 일어나면 안되기때문에 값으로 리턴한다.
func LoadRuleSetFromFile(filePath string) (RuleSet, error) {
	filePath, err := u.CheckPath(filePath)
	if err != nil {
		return RuleSet{}, err
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return RuleSet{}, fmt.Errorf("failed to access path: %w", err)
	}
	// filePath 가 디렉토리인지 확인
	if fileInfo.IsDir() {
		// 디렉토리 내 rule.json 파일 경로 확인
		filePath = filepath.Join(filePath, "rule.json")
	} else {
		return RuleSet{}, fmt.Errorf("path is not a directory: %s", filePath)
	}

	// rule.json 파일 존재 여부 확인
	bExist, _, err := u.FileExists(filePath)
	if !bExist {
		return RuleSet{}, fmt.Errorf("rule file not found at %s", filePath)
	}
	if err != nil {
		return RuleSet{}, fmt.Errorf("failed to access path: %w", err)
	}

	// utils 메서드로 대체, 주석 지우지 말것.
	/*if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return RuleSet{}, fmt.Errorf("rule file not found at %s", filePath)
	}*/

	// JSON 파일 읽기, 파일이 크지 않기 때문에 이렇게 처리 함.
	data, err := os.ReadFile(filePath)
	if err != nil {
		return RuleSet{}, fmt.Errorf("failed to read file: %w", err)
	}

	// JSON 데이터 디코딩
	var ruleSet RuleSet
	err = json.Unmarshal(data, &ruleSet)
	if err != nil {
		return RuleSet{}, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return ruleSet, nil
}

// Helper function: 파일 이름을 JSON 규칙의 delimiter 를 기준으로 파트로 나누기
func extractParts(fileName string, delimiters []string) []string {
	for _, delim := range delimiters {
		fileName = strings.ReplaceAll(fileName, delim, " ")
	}
	return strings.Fields(fileName)
}

// FilesToMap 파일 이름을 JSON 규칙에 따라 블록화하여 맵으로 변환
func FilesToMap(fileNames []string, ruleSet RuleSet) (map[int]map[string]string, error) {
	rowMap := make(map[string]int)               // Row Key → Row Index
	rowCounter := 0                              // 행 인덱스 카운터 (0부터 시작)
	resultMap := make(map[int]map[string]string) // 결과 데이터 저장용 맵

	for _, fileName := range fileNames {
		// 파일명을 JSON 규칙의 delimiter 를 기준으로 분리
		parts := extractParts(fileName, ruleSet.Delimiter)

		// Row Key 생성
		var rowKeyParts []string
		for _, idx := range ruleSet.RowRules.MatchParts {
			if idx < len(parts) {
				rowKeyParts = append(rowKeyParts, parts[idx])
			}
		}
		rowKey := strings.Join(rowKeyParts, "_")

		// Row Index 확인 및 추가
		if _, exists := rowMap[rowKey]; !exists {
			rowMap[rowKey] = rowCounter
			resultMap[rowCounter] = make(map[string]string)
			rowCounter++
		}

		// Column Key 생성 (ColumnRules.MatchParts 기준)
		var colKeyParts []string
		for _, idx := range ruleSet.ColumnRules.MatchParts {
			if idx < len(parts) {
				colKeyParts = append(colKeyParts, parts[idx])
			}
		}
		colKey := strings.Join(colKeyParts, "_")

		// Row 에 Column Key 와 파일명 추가
		rowIdx := rowMap[rowKey]
		resultMap[rowIdx][colKey] = fileName
	}

	return resultMap, nil
}

// FilterMap 컬럼 수를 검증하고 유효/무효 행을 분리하는 메서드
func FilterMap(resultMap map[int]map[string]string, expectedColCount int) (map[int]map[string]string, []map[string]string) {
	validRows := make(map[int]map[string]string)
	var invalidRows []map[string]string
	newRowCounter := 0

	for _, row := range resultMap {
		if len(row) == expectedColCount {
			validRows[newRowCounter] = row
			newRowCounter++
		} else {
			invalidRows = append(invalidRows, row)
		}
	}

	return validRows, invalidRows
}

// WriteInvalidFiles invalidRows 의 파일명을 하나의 텍스트 파일에 기록 TODO readonly 로 하는 것 생각
func WriteInvalidFiles(invalidRows []map[string]string, outputFilePath string) (err error) {
	// invalidRows 가 비어있으면 파일을 생성하지 않고 리턴
	if len(invalidRows) == 0 {
		return nil
	}

	// outputFilePath 가 디렉토리인지 확인
	fileInfo, err := os.Stat(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to access path %s: %w", outputFilePath, err)
	}

	// 디렉토리인 경우, 날짜와 시간을 포함한 파일명을 생성
	if fileInfo.IsDir() {
		timestamp := time.Now().Format("20060102150405") // 현재 날짜와 시간 (년월일시간분초)
		outputFilePath = filepath.Join(outputFilePath, fmt.Sprintf("invalid_files_%s.txt", timestamp))
	}

	// 출력 파일 생성 (덮어쓰기)
	file, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", outputFilePath, err)
	}

	defer func() {
		if cErr := file.Close(); cErr != nil {
			if err == nil {
				err = fmt.Errorf("failed to close file: %w", cErr)
			} else {
				err = fmt.Errorf("%v; failed to close file: %w", err, cErr)
			}
		}
	}()

	// 파일명들을 텍스트 파일에 기록
	for _, row := range invalidRows {
		for _, fileName := range row {
			_, err := file.WriteString(fileName + "\n")
			if err != nil {
				return fmt.Errorf("failed to write to file %s: %w", outputFilePath, err)
			}
		}
	}

	return err
}

// ValidateRuleSet validates the given rule set for conflicts and unused parts.
func ValidateRuleSet(ruleSet RuleSet) bool {
	hasConflict := false

	usageMap := make(map[int][]string)

	// Helper for registering usage of parts
	addUsage := func(indices []int, roleName string) {
		for _, idx := range indices {
			usageMap[idx] = append(usageMap[idx], roleName)
		}
	}

	// Register usages
	addUsage(ruleSet.RowRules.MatchParts, "RowRules.MatchParts")
	addUsage(ruleSet.ColumnRules.MatchParts, "ColumnRules.MatchParts")

	// Check for conflicts - any index used in more than one role is a conflict
	for idx, roles := range usageMap {
		if len(roles) > 1 {
			log.Printf("Conflict detected: part %d is used in multiple roles: %v", idx, roles)
			hasConflict = true
		}
	}

	return !hasConflict
}

// SaveResultMapToCSV map[int]map[string]string 데이터를 CSV 파일로 저장, TODO 파일 생성날짜를 기록할지 생각, readonly 로 하는 것 생각.
func SaveResultMapToCSV(resultMap map[int]map[string]string, headers []string, outputFilePath string) (err error) {
	outputFilePath, err = u.CheckPath(outputFilePath)
	if err != nil {
		return err
	}

	// filePath 가 디렉토리인지 확인
	fileInfo, err := os.Stat(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to access path: %w", err)
	}

	if fileInfo.IsDir() {
		// 디렉토리 경로에 fileblock.csv 파일 생성
		outputFilePath = filepath.Join(outputFilePath, "fileblock.csv")
	}

	file, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	//defer file.Close()
	defer func() {
		if cErr := file.Close(); cErr != nil {
			if err == nil {
				err = fmt.Errorf("failed to close file: %w", cErr)
			} else {
				err = fmt.Errorf("%v; failed to close file: %w", err, cErr)
			}
		}
	}()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 첫 번째 행에 헤더 추가
	headerRow := append([]string{"Row"}, headers...)
	if err := writer.Write(headerRow); err != nil {
		return fmt.Errorf("failed to write header row: %w", err)
	}

	var columnHeaders []string
	seen := make(map[string]struct{})

	// 모든 열 키를 동적으로 추출 (중복 제거)
	for _, row := range resultMap {
		for key := range row {
			if _, exists := seen[key]; !exists {
				columnHeaders = append(columnHeaders, key)
				seen[key] = struct{}{}
			}
		}
	}

	// 열 키를 정렬
	sort.Strings(columnHeaders)

	// 각 행 데이터를 CSV 에 추가
	for rowIdx := 0; rowIdx < len(resultMap); rowIdx++ {
		rowData := append([]string{fmt.Sprintf("Row%d", rowIdx)}, make([]string, len(columnHeaders))...)

		if row, exists := resultMap[rowIdx]; exists {
			for i, colKey := range columnHeaders {
				if value, ok := row[colKey]; ok {
					rowData[i+1] = value
				}
			}
		}

		if err := writer.Write(rowData); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return err
}

// GenerateMap 일단 이름 고침.
func GenerateMap(filePath string) (map[int]map[string]string, error) {
	// Load the rule set
	ruleSet, err := LoadRuleSetFromFile(filePath) // 이 메서드에서 filepath 의 검증을 해줌.
	if err != nil {
		return nil, fmt.Errorf("failed to load rule set: %w", err)
	}

	// Validate the rule set
	if !ValidateRuleSet(ruleSet) {
		return nil, fmt.Errorf("rule set has conflicts or unused parts")
	}

	// Read all file names from the directory
	// 예외 규정: rule.json, invalid_files 로 시작하는 파일, fileblock.csv
	exclusions := []string{"rule.json", "invalid_files", "fileblock.csv"}
	files, err := ReadAllFileNames(filePath, exclusions)

	if err != nil {
		return nil, fmt.Errorf("failed to read file names: %w", err)
	}

	resultMap, err := FilesToMap(files, ruleSet)
	if err != nil {
		return nil, fmt.Errorf("failed to blockify files: %w", err)
	}

	// Filter the result map into valid and invalid rows
	validRows, invalidRows := FilterMap(resultMap, len(ruleSet.Header))

	// Save valid rows to a CSV file
	if err := SaveResultMapToCSV(validRows, ruleSet.Header, filePath); err != nil {
		return nil, fmt.Errorf("failed to save result map to CSV: %w", err)
	}

	// Save invalid rows to a separate file
	if err := WriteInvalidFiles(invalidRows, filePath); err != nil {
		return nil, fmt.Errorf("failed to write invalid files: %w", err)
	}

	return validRows, nil
}

// GenerateFileBlock 일단 이름 고침. filePath 는 rule.josn 이 있는 위치이자 fileblock.csv, invalid_files, *.pb 파일 등이 가 저장될 위치.
func GenerateFileBlock(filePath string, files []string) (*pb.FileBlock, error) {
	// Load the rule set
	ruleSet, err := LoadRuleSetFromFile(filePath) // 이 메서드에서 filepath 의 검증을 해줌.
	if err != nil {
		return nil, fmt.Errorf("failed to load rule set: %w", err)
	}

	// Validate the rule set
	if !ValidateRuleSet(ruleSet) {
		return nil, fmt.Errorf("rule set has conflicts or unused parts")
	}

	resultMap, err := FilesToMap(files, ruleSet)
	if err != nil {
		return nil, fmt.Errorf("failed to blockify files: %w", err)
	}

	// Filter the result map into valid and invalid rows. 열의 갯수 기준으로 유효/무효 행을 분리
	validRows, invalidRows := FilterMap(resultMap, len(ruleSet.Header))

	// Save valid rows to a CSV file. 사용자에게 보여주기 위함.
	if err := SaveResultMapToCSV(validRows, ruleSet.Header, filePath); err != nil {
		return nil, fmt.Errorf("failed to save result map to CSV: %w", err)
	}

	// Save invalid rows to a separate file
	if err := WriteInvalidFiles(invalidRows, filePath); err != nil {
		return nil, fmt.Errorf("failed to write invalid files: %w", err)
	}

	// blockId 를 filePath 로 잡아둠.
	fbd := ConvertMapToFileBlock(validRows, ruleSet.Header, filePath)
	pbName := filepath.Join(filePath, fmt.Sprintf("%sfiles.pb", filepath.Base(filePath)))
	err = SaveProtoToFile(pbName, fbd, 0777)
	if err != nil {
		return nil, fmt.Errorf("failed to save proto to file: %w", err)
	}

	return fbd, nil
}

// ReadAllFileNames 디렉토리에서 파일을 읽되 예외 규정에 맞는 파일들은 제외
func ReadAllFileNames(dirPath string, exclusions []string) ([]string, error) {

	dirPath, err := u.CheckPath(dirPath)
	if err != nil {
		return nil, err
	}

	// 디렉토리의 파일 목록 읽기
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	// TODO 중복됨 중복되는 것 제거하는 방향으로.
	// excludeFiles 는 dirName 이 exclusions 목록에 있는 항목과 정확히 일치하거나,
	// 만약 exclusions 항목이 "*.확장자" 형태이면, dirName 에 해당 확장자가 포함되어 있으면 true 를 반환함.
	excludeFiles := func(fileName string, exclusions []string) bool {
		for _, ex := range exclusions {
			// 패턴이 "*.<ext>" 형식이면, 해당 확장자가 dirName 내에 존재하는지 확인함.
			if strings.HasPrefix(ex, "*.") {
				ext := ex[1:] // 예: "*.pb" -> ext 는 ".pb"
				if strings.Contains(fileName, ext) {
					return true
				}
			} else {
				// 일반적인 정확한 비교
				if fileName == ex {
					return true
				}
			}
		}
		return false
	}
	// 파일 이름을 저장할 슬라이스
	var fileNames []string
	// 파일 목록에서 제외할 파일들을 걸러내고 이름만 추출
	for _, file := range files {
		fileName := file.Name()

		// 예외 규정에 있는 파일이면 건너뛰기
		/*
			if u.ExcludeFiles(fileName, exclusions) {
				continue
			}
		*/
		if excludeFiles(fileName, exclusions) {
			continue
		}
		// 파일 이름을 경로와 함께 추가
		fileNames = append(fileNames, fileName)
	}

	return fileNames, nil
}

/*
// 기본 권한(0644) 사용
err := SaveProtoToFile("data.pb", message, 0644)

// 다른 권한 설정
err := SaveProtoToFile("data.pb", message, 0600) // 소유자만 읽기/쓰기 가능

// os.FileMode 상수 사용
err := SaveProtoToFile("data.pb", message, os.ModePerm) // 0777
*/

func SaveProtoToFile(filePath string, message proto.Message, perm os.FileMode) error {
	data, err := proto.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to serialize data: %w", err)
	}

	err = os.WriteFile(filePath, data, perm)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

func LoadFileBlock(filePath string) (*pb.FileBlock, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	message := &pb.FileBlock{}
	err = proto.Unmarshal(data, message)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize data: %w", err)
	}

	return message, nil
}

func LoadDataBlock(filePath string) (*pb.DataBlock, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	message := &pb.DataBlock{}
	err = proto.Unmarshal(data, message)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize data: %w", err)
	}

	return message, nil
}

// SaveFileBlockToTextFile 함수: FileBlock 를 텍스트 포맷으로 저장
func SaveFileBlockToTextFile(filePath string, data *pb.FileBlock) error {
	// proto 메시지를 텍스트 포맷으로 변환
	textData, err := prototext.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal to text format: %w", err)
	}

	// 텍스트 데이터를 파일에 저장
	return os.WriteFile(filePath, textData, os.ModePerm)
}

// MergeFileBlocks 여러 FileBlock 파일을 읽어 DataBlock 으로 합치는 메서드
func MergeFileBlocks(inputFiles []string, outputFile string) error {
	var blocks []*pb.FileBlock
	// 각 입력 파일을 로드하여 blocks 에 추가
	for _, file := range inputFiles {
		block, err := LoadFileBlock(file)
		if err != nil {
			return fmt.Errorf("failed to load file %s: %w", file, err)
		}
		blocks = append(blocks, block)
	}

	// DataBlockData 생성
	dataBlockData := &pb.DataBlock{
		UpdatedAt: timestamppb.Now(), // 현재 시간으로 설정
		Blocks:    blocks,
	}

	// DataBlockData 저장
	if err := SaveProtoToFile(outputFile, dataBlockData, os.ModePerm); err != nil {
		return fmt.Errorf("failed to save DataBlock: %w", err)
	}

	fmt.Printf("Successfully merged %d FileBlock files into %s\n", len(inputFiles), outputFile)
	return nil
}

// MergeFileBlocksFromData 여러 *pb.FileBlock 를 하나의 DataBlock 으로 통합
// 입력 파라미터가 이미 로드된 FileBlock 들의 슬라이스이므로, 별도의 파일 로딩 과정 없이 합친 결과를 반환함
func MergeFileBlocksFromData(inputBlocks []*pb.FileBlock) (*pb.DataBlock, error) {
	if len(inputBlocks) == 0 {
		return nil, fmt.Errorf("no input file blocks provided")
	}

	dataBlockData := &pb.DataBlock{
		UpdatedAt: timestamppb.Now(), // 현재 시간으로 설정
		Blocks:    inputBlocks,
	}

	// 합쳐진 결과를 반환 (필요하다면 로그 메시지 추가)
	fmt.Printf("Successfully merged %d FileBlockData into one DataBlockData\n", len(inputBlocks))
	return dataBlockData, nil
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

// GenerateRows 테스트 데이터 생성
func GenerateRows(data [][]string, headers []string) []*pb.Row {
	//rows := []*pb.Row{}
	rows := make([]*pb.Row, 0, len(data))
	for i, cells := range data {
		row := &pb.Row{
			RowNumber:   int32(i + 1), // 1부터 시작
			CellColumns: make(map[string]string, len(headers)),
		}
		for j, header := range headers {
			if j < len(cells) {
				row.CellColumns[header] = cells[j]
			}
		}
		rows = append(rows, row)
	}
	return rows
}

// ConvertMapToFileBlock map[int]map[string]string 를 FileBlockData 메시지로 변환
func ConvertMapToFileBlock(rows map[int]map[string]string, headers []string, blockID string) *pb.FileBlock {
	fbd := &pb.FileBlock{
		BlockId:       blockID,
		ColumnHeaders: headers, // 사용자 정의 헤더
		Rows:          make([]*pb.Row, 0, len(rows)),
	}

	// rowIndex 를 정렬해 순차적으로 처리
	rowIndices := make([]int, 0, len(rows))
	for idx := range rows {
		rowIndices = append(rowIndices, idx)
	}
	sort.Ints(rowIndices)

	for _, rIdx := range rowIndices {
		columns := rows[rIdx]
		r := &pb.Row{
			RowNumber:   int32(rIdx), // 1-based. 필요에 맞게 조정
			CellColumns: make(map[string]string, len(columns)),
		}
		// 열 데이터를 그대로 저장
		for colKey, value := range columns {
			r.CellColumns[colKey] = value
		}
		fbd.Rows = append(fbd.Rows, r)
	}
	return fbd
}
