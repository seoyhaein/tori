package service

import (
	"context"
	"database/sql"
	"fmt"
	pb "github.com/seoyhaein/api-protos/gen/go/datablock/ichthys"
	"github.com/seoyhaein/api-protos/gen/go/datablock/ichthys/service"
	"github.com/seoyhaein/tori/config"
	dbUtils "github.com/seoyhaein/tori/db"
	globallog "github.com/seoyhaein/tori/log"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/types/known/timestamppb"
	"os"
	"path/filepath"
)

var logger = globallog.Log

// DataBlockCliService encapsulates core folder/database operations for CLI and gRPC.
type DataBlockCliService struct {
	db  *sql.DB
	cfg *config.Config
}

// NewDataBlockCliService constructs a new CLI service instance.
func NewDataBlockCliService(dbConn *sql.DB, cfg *config.Config) *DataBlockCliService {
	return &DataBlockCliService{db: dbConn, cfg: cfg}
}

// GetDataBlock loads the DataBlock and applies timestamp-based logic.
func (s *DataBlockCliService) GetDataBlock(ctx context.Context, updateAt *timestamppb.Timestamp) (*pb.DataBlock, error) {
	// 서버의 데이터 블록 경로 정리
	dataBlockPath := filepath.Clean(s.cfg.RootDir)
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

// SaveFolders 폴더 정보를 DB에 저장, TODO 이건 한번만 실행되어야 하는 메서드 임. 이름을 이러한 맥락을 고려해서 넣어 주어야 할듯
func (s *DataBlockCliService) SaveFolders(ctx context.Context) error {
	err := dbUtils.SaveFolders(ctx, s.db, s.cfg.RootDir, nil, s.cfg.FilesExclusions)
	return err
}

func (s *DataBlockCliService) SyncFolders(ctx context.Context) (bool, error) {
	// 디렉터리 경로와 파일 제외 패턴을 넘겨서 dbUtils 쪽으로 위임
	return dbUtils.SyncFolders(ctx, s.db, s.cfg.RootDir, nil, s.cfg.FilesExclusions)
}

// DataBlockServer bridges DataBlockCliService with the gRPC interface.
type DataBlockServer struct {
	pb.UnimplementedDataBlockServiceServer
	core *DataBlockCliService
}

// NewDataBlockServer wraps DataBlockCliService for gRPC.
func NewDataBlockServer(core *DataBlockCliService) pb.DataBlockServiceServer {
	return &DataBlockServer{core: core}
}

// SyncFolders RPC handler.
/*func (s *DataBlockServer) SyncFolders(ctx context.Context, _ *emptypb.Empty) (*pb.SyncResponse, error) {
	updated, err := s.core.SyncFolders(ctx)
	return &pb.SyncResponse{Updated: updated}, err
}*/

// SaveFolders RPC handler.
/*func (s *DataBlockServer) SaveFolders(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.core.SaveFolders(ctx)
}*/

// GetDataBlock RPC handler.
/*func (s *DataBlockServer) GetDataBlock(ctx context.Context, req *pb.DataBlockRequest) (*pb.DataBlock, error) {
	return s.core.GetDataBlock(ctx, req)
}*/

// TODO 이건 api-proto 프로젝트로 빼자.

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
