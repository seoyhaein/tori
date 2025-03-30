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

	// TODO 이 구문들이 반복적으로 사용되는데, 이거 한번에 사용하도록 하자. config 같은 경우도 한번에 처리하는 방향으로 가자.
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
