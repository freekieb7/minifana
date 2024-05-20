package main

import (
	"context"
	"github.com/tidwall/wal"
	v1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type LogsService interface {
	v1.LogsServiceServer
	Register(grpcServer *grpc.Server)
}

type logsService struct {
	v1.UnimplementedLogsServiceServer
}

func NewLogsService() LogsService {
	return &logsService{}
}

func (s *logsService) Register(grpcServer *grpc.Server) {
	v1.RegisterLogsServiceServer(grpcServer, s)
}

func (s *logsService) Export(_ context.Context, req *v1.ExportLogsServiceRequest) (*v1.ExportLogsServiceResponse, error) {
	log, err := wal.Open("/minifana/logs/wal", nil)

	// write some entries
	li, _ := log.LastIndex()
	res, _ := proto.Marshal(req)

	err = log.Write(li+1, res)

	// close the log
	err = log.Close()

	return new(v1.ExportLogsServiceResponse), err
}
