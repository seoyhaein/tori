package service

import (
	"context"
	"github.com/seoyhaein/tori/api"
	pb "github.com/seoyhaein/tori/protos"
	"google.golang.org/grpc"
)

type dBApisServiceServerImpl struct {
	pb.UnimplementedDBApisServiceServer
	dbApis api.DBApis
}

func NewDBApisServiceServer() pb.DBApisServiceServer {
	return &dBApisServiceServerImpl{
		dbApis: api.NewDBApis(),
	}
}

func RegisterDBApisServiceServer(service *grpc.Server) {
	pb.RegisterDBApisServiceServer(service, NewDBApisServiceServer())
}

// SyncFoldersInfo TODO 여기서 update 날짜 보내는데 어떻게 되는지 확인해봐야 함. 사릴 업데이트가 되면 datablock 도 가져와야 할듯. 이건 force 설정과 관련해서 잘 생각해야 함.
// 에러가 발생하면 nil 을 주자. in 지금은 사용하지 않음. 확인해야함. GetDataBlock 과 연계해서 잘 생각해야 함.
func (s *dBApisServiceServerImpl) SyncFoldersInfo(ctx context.Context, in *pb.SyncFoldersInfoRequest) (*pb.SyncFoldersInfoResponse, error) {

	isUpdated, err := s.dbApis.SyncFoldersInfo(ctx)
	if err != nil {
		return nil, err
	}
	if isUpdated {
		return &pb.SyncFoldersInfoResponse{
			Updated: true,
		}, nil
	}
	return &pb.SyncFoldersInfoResponse{
		Updated: false,
	}, nil
}
