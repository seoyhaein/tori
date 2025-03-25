package v1rpc

import (
	pb "github.com/seoyhaein/tori/protos"
	"google.golang.org/protobuf/proto"
	"os"
	"testing"
)

// TODO 테스트 코드 완성해야함.

// TestSaveProtoToFile 는 SaveProtoToFile 함수를 테스트함.
func TestSaveProtoToFile(t *testing.T) {
	// 임시 파일 생성. 파일은 테스트 후 삭제됨.
	tmpFile, err := os.CreateTemp("", "testproto_*.bin")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}
	tmpFilePath := tmpFile.Name()
	// 파일 핸들은 SaveProtoToFile 함수에서 파일을 쓰기 때문에 닫고,
	// 테스트 종료 시 파일 삭제
	cErr := tmpFile.Close()
	if cErr != nil {
		t.Logf("failed to close temp file: %v", cErr)
	}

	defer func() {
		rErr := os.Remove(tmpFilePath)
		if rErr != nil {
			t.Logf("failed to remove temp file: %v", rErr)
		}
	}()

	// 테스트용 proto 메시지 생성. 여기서는 StoreFoldersInfoRequest 를 사용함.
	originalMsg := &pb.StoreFoldersInfoRequest{
		Confirm: true,
	}

	// SaveProtoToFile 호출: proto 메시지를 임시 파일에 저장
	err = SaveProtoToFile(tmpFilePath, originalMsg, 0644)
	if err != nil {
		t.Fatalf("SaveProtoToFile failed: %v", err)
	}

	// 파일에서 데이터 읽어오기
	data, err := os.ReadFile(tmpFilePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	// 읽어온 데이터를 새 proto 메시지로 Unmarshal
	var newMsg pb.StoreFoldersInfoRequest
	if err := proto.Unmarshal(data, &newMsg); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	// 원본 메시지와 새 메시지의 필드가 동일한지 확인
	if originalMsg.Confirm != newMsg.Confirm {
		t.Errorf("message mismatch: expected Confirm=%v, got Confirm=%v", originalMsg.Confirm, newMsg.Confirm)
	}
}

// TestLoadFileBlock 는 임시 파일에 저장된 FileBlockData 메시지를 읽어오는 함수를 테스트함.
func TestLoadFileBlock(t *testing.T) {
	// 임시 파일 생성
	tmpFile, err := os.CreateTemp("", "fileblock_*.bin")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}
	tmpFilePath := tmpFile.Name()

	// 파일 핸들을 닫는다.
	if cErr := tmpFile.Close(); cErr != nil {
		t.Logf("failed to close temp file: %v", cErr)
	}

	// 테스트 종료 시 임시 파일 삭제
	defer func() {
		if rErr := os.Remove(tmpFilePath); rErr != nil {
			t.Logf("failed to remove temp file: %v", rErr)
		}
	}()

	// 테스트용 FileBlockData 메시지 생성
	originalMsg := &pb.FileBlock{
		BlockId:       "test-block",
		ColumnHeaders: []string{"col1", "col2"},
		Rows: []*pb.Row{
			{
				RowNumber: 1,
				CellColumns: map[string]string{
					"col1": "value1",
					"col2": "value2",
				},
			},
		},
	}

	// 메시지를 직렬화 (marshal)
	data, err := proto.Marshal(originalMsg)
	if err != nil {
		t.Fatalf("failed to marshal FileBlockData: %v", err)
	}

	// 임시 파일에 직렬화한 데이터를 기록
	if err := os.WriteFile(tmpFilePath, data, 0644); err != nil {
		t.Fatalf("failed to write to temporary file: %v", err)
	}

	// LoadFileBlock 함수를 사용하여 파일에서 메시지를 읽어옴
	loadedMsg, err := LoadFileBlock(tmpFilePath)
	if err != nil {
		t.Fatalf("LoadFileBlock failed: %v", err)
	}

	// 원본 메시지와 로드한 메시지가 동일한지 비교
	if !proto.Equal(originalMsg, loadedMsg) {
		t.Errorf("Loaded FileBlockData does not match original.\nOriginal: %v\nLoaded: %v", originalMsg, loadedMsg)
	}
}
