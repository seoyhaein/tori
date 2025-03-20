package v1rpc

import (
	"fmt"
	pb "github.com/seoyhaein/tori/protos"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"os"
	"sort"
)

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

/*
// 기본 권한(0644) 사용
err := SaveProtoToFile("data.pb", message, 0644)

// 다른 권한 설정
err := SaveProtoToFile("data.pb", message, 0600) // 소유자만 읽기/쓰기 가능

// os.FileMode 상수 사용
err := SaveProtoToFile("data.pb", message, os.ModePerm) // 0777
*/

func LoadFileBlock(filePath string) (*pb.FileBlockData, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	message := &pb.FileBlockData{}
	err = proto.Unmarshal(data, message)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize data: %w", err)
	}

	return message, nil
}

func LoadDataBlock(filePath string) (*pb.DataBlockData, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	message := &pb.DataBlockData{}
	err = proto.Unmarshal(data, message)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize data: %w", err)
	}

	return message, nil
}

// SaveFileBlockToTextFile 함수: FileBlock 를 텍스트 포맷으로 저장
func SaveFileBlockToTextFile(filePath string, data *pb.FileBlockData) error {
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
	var blocks []*pb.FileBlockData
	// 각 입력 파일을 로드하여 blocks 에 추가
	for _, file := range inputFiles {
		block, err := LoadFileBlock(file)
		if err != nil {
			return fmt.Errorf("failed to load file %s: %w", file, err)
		}
		blocks = append(blocks, block)
	}

	// DataBlockData 생성
	dataBlockData := &pb.DataBlockData{
		Blocks: blocks,
	}

	// DataBlockData 저장
	if err := SaveProtoToFile(outputFile, dataBlockData, os.ModePerm); err != nil {
		return fmt.Errorf("failed to save DataBlock: %w", err)
	}

	fmt.Printf("Successfully merged %d FileBlock files into %s\n", len(inputFiles), outputFile)
	return nil
}

// MergeFileBlocksFromData 여러 *pb.FileBlockData 를 하나의 DataBlockData 으로 통합
// 입력 파라미터가 이미 로드된 FileBlockData 들의 슬라이스이므로, 별도의 파일 로딩 과정 없이 합친 결과를 반환함
func MergeFileBlocksFromData(inputBlocks []*pb.FileBlockData) (*pb.DataBlockData, error) {
	if len(inputBlocks) == 0 {
		return nil, fmt.Errorf("no input file blocks provided")
	}

	dataBlockData := &pb.DataBlockData{
		UpdatedAt: timestamppb.Now(), // 현재 시간으로 설정
		Blocks:    inputBlocks,
	}

	// 합쳐진 결과를 반환 (필요하다면 로그 메시지 추가)
	fmt.Printf("Successfully merged %d FileBlockData into one DataBlockData\n", len(inputBlocks))
	return dataBlockData, nil
}

// SaveDataBlockToTextFile DataBlockData 텍스트 포맷으로 파일에 저장
func SaveDataBlockToTextFile(filePath string, data *pb.DataBlockData) error {
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

// ConvertMapToFileBlockData map[int]map[string]string 를 FileBlockData 메시지로 변환 TODO 에러 넣어야 하지 않을까?
func ConvertMapToFileBlockData(rows map[int]map[string]string, headers []string, blockID string) *pb.FileBlockData {
	fbd := &pb.FileBlockData{
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
