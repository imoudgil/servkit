// Package timegrpc implements the generated TimeService gRPC API.
package timegrpc

import (
	"context"
	"time"

	timev1 "github.com/imoudgil/servkit/proto/time/v1"
)

// Server implements time.v1.TimeService.
type Server struct {
	timev1.UnimplementedTimeServiceServer
	ServiceName string
}

// Now returns the current unix timestamp for the configured service name.
func (s *Server) Now(_ context.Context, _ *timev1.NowRequest) (*timev1.NowResponse, error) {
	name := s.ServiceName
	if name == "" {
		name = "servkit"
	}
	return &timev1.NowResponse{
		Unix:    time.Now().Unix(),
		Service: name,
	}, nil
}
