package main

import (
	"context"
	"github.com/tidwall/wal"
	v1 "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type TraceService interface {
	v1.TraceServiceServer
	Register(grpcServer *grpc.Server)
}

type traceService struct {
	v1.UnimplementedTraceServiceServer
}

func NewTraceService() TraceService {
	return &traceService{}
}

func (s *traceService) Register(grpcServer *grpc.Server) {
	v1.RegisterTraceServiceServer(grpcServer, s)
}

func (s *traceService) Export(_ context.Context, req *v1.ExportTraceServiceRequest) (*v1.ExportTraceServiceResponse, error) {
	// open a new log file
	log, err := wal.Open("/minifana/trace/wal", nil)

	// write some entries
	li, _ := log.LastIndex()
	res, _ := proto.Marshal(req)

	err = log.Write(li+1, res)

	// close the log
	err = log.Close()

	return new(v1.ExportTraceServiceResponse), err
}
