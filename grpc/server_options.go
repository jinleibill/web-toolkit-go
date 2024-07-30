package grpc

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type ServerConfig []grpc.ServerOption

type ServerOption func(*ServerConfig)

func WithServerOptionTLS(certPath string, keyPath string) ServerOption {
	return func(config *ServerConfig) {
		cred, err := credentials.NewServerTLSFromFile(certPath, keyPath)
		if err != nil {
			panic(err)
		}

		*config = append(*config, grpc.Creds(cred))
	}
}

func WithUnaryServerInterceptors(interceptors ...grpc.UnaryServerInterceptor) ServerOption {
	return func(config *ServerConfig) {
		unary := grpc.ChainUnaryInterceptor(interceptors...)
		*config = append(*config, unary)
	}
}

func WithStreamServerInterceptors(interceptors ...grpc.StreamServerInterceptor) ServerOption {
	return func(config *ServerConfig) {
		stream := grpc.ChainStreamInterceptor(interceptors...)
		*config = append(*config, stream)
	}
}
