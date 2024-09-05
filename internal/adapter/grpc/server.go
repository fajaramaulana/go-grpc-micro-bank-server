package grpc

import (
	"fmt"
	"net"

	"github.com/fajaramaulana/go-grpc-micro-bank-proto/protogen/go/bank"
	"github.com/fajaramaulana/go-grpc-micro-bank-proto/protogen/go/resilliency"
	"github.com/fajaramaulana/go-grpc-micro-bank-server/internal/logger"
	"github.com/fajaramaulana/go-grpc-micro-bank-server/internal/port"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GrpcAdapter struct {
	bankService        port.BankServicePort
	resilliencyService port.ResilliencyServicePort
	grpcPort           int
	server             *grpc.Server
	bank.BankServiceServer
	resilliency.ResilliencyServiceServer
}

func NewGrpcAdapter(bankService port.BankServicePort, resilliencyService port.ResilliencyServicePort, grpcPort int) *GrpcAdapter {
	return &GrpcAdapter{
		bankService:        bankService,
		grpcPort:           grpcPort,
		resilliencyService: resilliencyService,
	}
}

func (a *GrpcAdapter) Run() {
	var err error

	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", a.grpcPort))

	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to listen on port %d", a.grpcPort)
	}

	log.Info().Msgf("Server listening on port %d", a.grpcPort)

	grpcLogger := grpc.UnaryInterceptor(logger.GrpcLogger)
	grpcServer := grpc.NewServer(grpcLogger)
	reflection.Register(grpcServer)
	a.server = grpcServer

	bank.RegisterBankServiceServer(grpcServer, a)
	resilliency.RegisterResilliencyServiceServer(grpcServer, a)

	if err := grpcServer.Serve(listen); err != nil {
		log.Fatal().Err(err).Msgf("Failed to serve gRPC server over port %d", a.grpcPort)
	}
}

func (a *GrpcAdapter) Stop() {
	a.server.Stop()
}
