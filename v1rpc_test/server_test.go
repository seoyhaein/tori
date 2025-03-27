package v1rpc_test

import (
	"context"
	globallog "github.com/seoyhaein/tori/log"
	"github.com/seoyhaein/tori/v1rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"os"
	"syscall"
	"testing"
	"time"
)

var Log = globallog.Log

func TestServerHealth(t *testing.T) {
	v1rpc.Address = "localhost:50053"

	// gRPC 서버를 별도 고루틴에서 실행하고, 종료 에러를 받을 채널 생성
	serverErrCh := make(chan error, 1)
	go func() {
		err := v1rpc.Server()
		serverErrCh <- err
	}()

	// 서버가 시작될 시간을 잠시 대기
	time.Sleep(200 * time.Millisecond)

	// 서버에 연결 (grpc.DialContext 사용)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, v1rpc.Address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer func(conn *grpc.ClientConn) {
		cErr := conn.Close()
		if cErr != nil {
			Log.Warnf("Failed to close gRPC connection: %v", cErr)
		}
	}(conn)

	// 헬스 체크 클라이언트 생성
	healthClient := grpc_health_v1.NewHealthClient(conn)
	healthResp, err := healthClient.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
	if healthResp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		t.Errorf("Health status mismatch. Expected: %v, Got: %v", grpc_health_v1.HealthCheckResponse_SERVING, healthResp.Status)
	}

	// 서버 graceful shutdown 테스트를 위해 SIGINT 신호 전송
	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("Failed to find process: %v", err)
	}
	err = proc.Signal(syscall.SIGINT)
	if err != nil {
		t.Fatalf("Failed to send SIGINT: %v", err)
	}

	// 서버 종료를 기다림
	err = <-serverErrCh
	if err != nil {
		t.Errorf("Server shutdown returned error: %v", err)
	}
}
