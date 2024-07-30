package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

type Server interface {
	grpc.ServiceRegistrar
	Start() error
	Shutdown(ctx context.Context) error
}

var _ Server = (*server)(nil)

type server struct {
	s       *grpc.Server
	address string
}

func NewServer(address string, options ...ServerOption) Server {
	cfg := ServerConfig{}

	for _, option := range options {
		option(&cfg)
	}

	rpcServer := grpc.NewServer(cfg...)

	reflection.Register(rpcServer)

	return &server{
		s:       rpcServer,
		address: address,
	}
}

func (s *server) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	s.s.RegisterService(desc, impl)
}

func (s *server) Start() error {
	var err error
	listen, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("failed to listen on address: %v, error: %v", s.address, err)
	}

	return s.s.Serve(listen)
}

func (s *server) Shutdown(ctx context.Context) error {
	stopped := make(chan struct{})
	go func() {
		s.s.GracefulStop()
		close(stopped)
	}()

	select {
	case <-ctx.Done():
		s.s.Stop()
		return fmt.Errorf("grpc server failed to stop gracefully")
	case <-stopped:
		return nil
	}
}
