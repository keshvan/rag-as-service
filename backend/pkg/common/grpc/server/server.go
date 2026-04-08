package server

import (
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/keshvan/rag-as-service/backend/pkg/common/config"
	"github.com/keshvan/rag-as-service/backend/pkg/common/grpc/interceptors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GRPCServer struct {
	server *grpc.Server
	cfg    config.GRPCServerConfig
	log    *slog.Logger
}

func NewGRPCServer(cfg config.GRPCServerConfig) *GRPCServer {
	s := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.ServerTenantInterceptor),
	)
	reflection.Register(s)
	return &GRPCServer{server: s, cfg: cfg, log: slog.Default()}
}

func (s *GRPCServer) Run(registerFunc func(*grpc.Server)) error {
	addr := s.cfg.GetAddr()
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	registerFunc(s.server)

	go s.gracefulShutdown()

	s.log.Info("gRPC server started", "addr", addr)

	return s.server.Serve(lis)
}

func (s *GRPCServer) gracefulShutdown() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	s.log.Info("shutting down gRPC server")

	done := make(chan struct{})

	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		s.log.Info("gRPC server stopped gracefully")
	case <-time.After(5 * time.Second):
		s.log.Warn("forcing gRPC stop")
		s.server.Stop()
	}
}
