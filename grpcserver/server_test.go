package grpcserver

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/imoudgil/servkit/auth"
	"github.com/imoudgil/servkit/config"
	timev1 "github.com/imoudgil/servkit/proto/time/v1"
	"github.com/imoudgil/servkit/timegrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

func TestGRPCServerNowRPC(t *testing.T) {
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)

	validator := auth.New("test-token")
	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(validator.UnaryServerInterceptor()))
	timev1.RegisterTimeServiceServer(srv, &timegrpc.Server{ServiceName: "servkit"})

	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	client := timev1.NewTimeServiceClient(conn)
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer test-token"))
	resp, err := client.Now(ctx, &timev1.NowRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Service != "servkit" {
		t.Fatalf("service = %q", resp.Service)
	}
}

func TestListenAndServeShutdown(t *testing.T) {
	t.Setenv("GRPC_ADDR", "127.0.0.1:0")
	t.Setenv("SHUTDOWN_TIMEOUT", "2s")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}

	s, err := New(cfg, func(gs *grpc.Server) {
		timev1.RegisterTimeServiceServer(gs, &timegrpc.Server{ServiceName: cfg.Name})
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.ListenAndServe(ctx) }()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil && err != context.Canceled {
			t.Fatalf("ListenAndServe() = %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("server did not shut down in time")
	}
}
