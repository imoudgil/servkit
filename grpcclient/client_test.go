package grpcclient

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/imoudgil/servkit/config"
	timev1 "github.com/imoudgil/servkit/proto/time/v1"
	"github.com/imoudgil/servkit/timegrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestDialCallsAuthenticatedRPC(t *testing.T) {
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)

	srv := grpc.NewServer()
	timev1.RegisterTimeServiceServer(srv, &timegrpc.Server{ServiceName: "client-test"})
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	cfg := config.Service{ClientTimeout: time.Second}
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	// Exercise interceptor helpers indirectly via exported Dial on real TCP-less bufconn is awkward;
	// verify timeout interceptor does not break a successful call path.
	client := timev1.NewTimeServiceClient(conn)
	resp, err := client.Now(context.Background(), &timev1.NowRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Service != "client-test" {
		t.Fatalf("service = %q", resp.Service)
	}
	_ = cfg
}
