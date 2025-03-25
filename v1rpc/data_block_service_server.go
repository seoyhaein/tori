package v1rpc

import (
	"context"
	pb "github.com/seoyhaein/tori/protos"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
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
	// TODO 아래 내용 구현 해야 함. 테스트 코드오 이렇게 나오도록 했음. 테스트 코드도 수정 필요함.
	// 예시: 현재 시간을 updated_at 으로 설정하고, 빈 FileBlockData 리스트를 포함하는 DataBlockData 생성
	dataBlock := &pb.DataBlock{
		UpdatedAt: timestamppb.Now(),
		Blocks:    []*pb.FileBlock{}, // 실제 구현 시, 실제 파일 블록 데이터를 채워 넣어야 함.
	}

	return &pb.GetDataBlockResponse{
		Data: dataBlock,
	}, nil
}
