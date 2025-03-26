package api

import (
	"context"
	pb "github.com/seoyhaein/tori/protos"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func GetDataBlock(ctx context.Context, req *pb.GetDataBlockRequest) (*timestamppb.Timestamp, error) {
	return nil, nil
}

func SyncFoldersInfo(ctx context.Context, force bool) (bool, error) {
	return false, nil
}
