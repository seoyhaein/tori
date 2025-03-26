package v1rpc

import (
	"context"
	pb "github.com/seoyhaein/tori/protos"
	"google.golang.org/grpc"
)

type dBApisServiceServerImpl struct {
	pb.UnimplementedDBApisServiceServer
}

func NewDBApisServiceServer() pb.DBApisServiceServer {
	return &dBApisServiceServerImpl{}
}

func RegisterDBApisServiceServer(service *grpc.Server) {
	pb.RegisterDBApisServiceServer(service, NewDBApisServiceServer())
}

func (s *dBApisServiceServerImpl) SyncFoldersInfo(ctx context.Context, in *pb.SyncFoldersInfoRequest) (*pb.SyncFoldersInfoResponse, error) {

	return &pb.SyncFoldersInfoResponse{}, nil
}
