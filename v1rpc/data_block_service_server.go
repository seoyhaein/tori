package v1rpc

import (
	"context"
	"github.com/seoyhaein/tori/config"
	pb "github.com/seoyhaein/tori/protos"
	"google.golang.org/grpc"
	"path/filepath"
)

type dataBlockServiceServerImpl struct {
	pb.UnimplementedDataBlockServiceServer
}

// NewDataBlockServiceServer 는 DataBlockServiceServer 의 새로운 인스턴스를 반환
func NewDataBlockServiceServer() pb.DataBlockServiceServer {
	return &dataBlockServiceServerImpl{}
}

// RegisterDataBlockServiceServer server.go 에서 서비스 등록할때 사용하면 서비스 사용가능.
func RegisterDataBlockServiceServer(service *grpc.Server) {
	pb.RegisterDataBlockServiceServer(service, NewDataBlockServiceServer())
}

// GetDataBlock 는 클라이언트의 빈 요청에 대해 DataBlock 를 응답으로 반환
func (s *dataBlockServiceServerImpl) GetDataBlock(ctx context.Context, in *pb.GetDataBlockRequest) (*pb.GetDataBlockResponse, error) {
	// TODO 수정해줘야 함.
	var dataBlockPath string
	if config.GlobalConfig == nil {
		dataBlockPath = "/test/datablock.pb"
	} else {
		dataBlockPath = filepath.Join(config.GlobalConfig.RootDir, "datablock.pb")
	}

	dataBlock, err := LoadDataBlock(dataBlockPath)

	if err != nil {
		return nil, err
	}

	return &pb.GetDataBlockResponse{
		Data: dataBlock,
	}, nil
}
