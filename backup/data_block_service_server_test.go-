package v1rpc_test

import (
	"context"
	pb "github.com/seoyhaein/tori/protos"
	"github.com/seoyhaein/tori/v1rpc/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"testing"
	"time"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener
var serverErr error

// 인메모리 gRPC 서버 설정
func init() {
	lis = bufconn.Listen(bufSize)
	grpcServer := grpc.NewServer()
	// DataBlockServiceServer 구현체를 등록 (구현체는 NewDataBlockServiceServer()로 생성)
	// 테스트의 독립성을 위해서 이거 사용하지 않음.
	// v1rpc.RegisterDataBlockServiceServer(grpcServer)
	pb.RegisterDataBlockServiceServer(grpcServer, service.NewDataBlockServiceServer())
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			serverErr = err
		}
	}()
}

// bufDialer 는 bufconn 을 통해 연결을 생성함.
func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestGetDataBlockData(t *testing.T) {
	// 서버 시작 에러 확인
	if serverErr != nil {
		t.Fatalf("Failed to start gRPC server: %v", serverErr)
	}

	ctx := context.Background()
	// bufconn 을 통한 인메모리 연결, DialContext 는 bufconn 테스트에 적합 따라서 DialContext 와 WithInsecure 이 메서드들이 deprecated api 이지만 그대로 사용.
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to connect via bufnet: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Logf("Failed to close connection: %v", err)
		}
	}()

	// 생성된 proto 클라이언트 스텁 사용
	client := pb.NewDataBlockServiceClient(conn)
	// 일단 DataBlock 이 없다고 할때를 기준으로 잡음. 무조건 데이터를 전송받게 됨.
	// TODO 이후 CurrentUpdateAt 에 따라서 테스트 진행해야함.
	resp, err := client.GetDataBlock(ctx, &pb.GetDataBlockRequest{})
	if err != nil {
		t.Fatalf("GetDataBlockData call failed: %v", err)
	}

	// 응답 메시지의 Data 필드 검사
	if resp.Data == nil {
		t.Error("Response data is nil.")
	} else {
		// updated_at 필드가 설정되어 있는지 확인
		if updatedAt := resp.Data.GetUpdatedAt(); updatedAt == nil {
			t.Error("Response UpdatedAt field is nil.")
		} else {
			t.Logf("UpdatedAt: %v", updatedAt)
		}
		// 파일 블럭 리스트 결과를 로그로 출력 (조건 검사는 하지 않음)
		t.Logf("File blocks: %+v", resp.Data.GetBlocks())
		t.Logf("Response Data: %+v", resp.Data)
	}

	// 테스트가 종료되기를 충분히 기다리도록 잠깐 sleep
	time.Sleep(100 * time.Millisecond)
}
