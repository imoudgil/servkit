package timegrpc

import (
	"context"
	"testing"

	timev1 "github.com/imoudgil/servkit/proto/time/v1"
)

func TestNowReturnsServiceName(t *testing.T) {
	s := &Server{ServiceName: "demo-api"}
	resp, err := s.Now(context.Background(), &timev1.NowRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Service != "demo-api" {
		t.Fatalf("service = %q", resp.Service)
	}
	if resp.Unix <= 0 {
		t.Fatalf("unix = %d", resp.Unix)
	}
}
