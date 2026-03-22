package server

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/keshvan/rag-as-service/backend/pkg/common/config"
	"github.com/keshvan/rag-as-service/backend/pkg/common/grpc/interceptors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GRPCServer struct {
	server *grpc.Server
	cfg    config.GRPCServerConfig
}

func NewGRPCServer(cfg config.GRPCServerConfig) *GRPCServer {
	s := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.ServerTenantInterceptor),
	)
	reflection.Register(s)
	return &GRPCServer{server: s, cfg: cfg}
}

func (s *GRPCServer) Run(registerFunc func(*grpc.Server)) error {
	addr := s.cfg.GetAddr()
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	registerFunc(s.server)

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop

		log.Printf("Shutting down gRPC server on %s...", addr)
		s.server.GracefulStop()
	}()

	log.Printf("gRPC server started on %s", addr)
	return s.server.Serve(lis)
}
