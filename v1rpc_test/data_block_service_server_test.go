package v1rpc_test

import (
	"context"
	pb "github.com/seoyhaein/tori/protos"
	"github.com/seoyhaein/tori/v1rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
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
	pb.RegisterDataBlockServiceServer(grpcServer, v1rpc.NewDataBlockServiceServer())
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

	// 빈 요청(emptypb.Empty)으로 GetDataBlockData 호출
	resp, err := client.GetDataBlockData(ctx, &emptypb.Empty{})
	if err != nil {
		t.Fatalf("GetDataBlockData call failed: %v", err)
	}

	// 응답 메시지의 Data 필드 검사
	if resp.Data == nil {
		t.Error("Response data is nil.")
	} else {
		// updated_at 필드가 설정되어 있는지 확인
		if resp.Data.GetUpdatedAt() == nil {
			t.Error("Response UpdatedAt field is nil.")
		}
		// 파일 블럭 리스트가 비어 있어야 함 (예시)
		if len(resp.Data.GetBlocks()) != 0 {
			t.Errorf("File block count mismatch. Expected: %d, Got: %d", 0, len(resp.Data.GetBlocks()))
		}
		t.Logf("Response Data: %+v", resp.Data)
	}

	// 테스트가 충분히 기다리도록 잠깐 sleep (예시)
	time.Sleep(100 * time.Millisecond)
}
