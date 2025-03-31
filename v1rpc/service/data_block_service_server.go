package service

import (
	"context"
	"github.com/seoyhaein/tori/api"
	pb "github.com/seoyhaein/tori/protos"
	"google.golang.org/grpc"
)

//var Config = c.GlobalConfig

type dataBlockServiceServerImpl struct {
	pb.UnimplementedDataBlockServiceServer
	dbApis api.DBApis
}

// TODO 중요. 이거 반드시 처리 해야함.
// TODO NewDBApis 는 main 의 init 에서 config 처리를 함으로 테스트의 경우 에러 날 수 있음. 이거 보안하는 방향으로 처리 해야함.
// TODO init 처리 할때 main 에서만 처리하지 말고 다른 패키지에서 처리하는 방향을 생각해봐.

// NewDataBlockServiceServer 는 DataBlockServiceServer 의 새로운 인스턴스를 반환
func NewDataBlockServiceServer() pb.DataBlockServiceServer {
	return &dataBlockServiceServerImpl{
		dbApis: api.NewDBApis(),
	}
}

// RegisterDataBlockServiceServer server.go 에서 서비스 등록할때 사용하면 서비스 사용가능.
func RegisterDataBlockServiceServer(service *grpc.Server) {
	pb.RegisterDataBlockServiceServer(service, NewDataBlockServiceServer())
}

// GetDataBlock 는 클라이언트의 빈 요청에 또는 버전 요청에 대해서 DataBlock 또는 nil 을 반환한다.
func (s *dataBlockServiceServerImpl) GetDataBlock(ctx context.Context, in *pb.GetDataBlockRequest) (*pb.GetDataBlockResponse, error) {

	dataBlock, err := s.dbApis.GetDataBlock(ctx, in.CurrentUpdatedAt)
	if err != nil {
		return nil, err
	}
	// 클라이언트와 서버의 버전이 동일하다면 업데이트 할 필요 없음.
	if dataBlock == nil {
		return &pb.GetDataBlockResponse{
			Data:     nil,
			NoUpdate: true,
		}, nil
	}
	// 최신 버전 업데이트.
	return &pb.GetDataBlockResponse{
		Data:     dataBlock,
		NoUpdate: false,
	}, nil
}
