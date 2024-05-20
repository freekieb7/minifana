package main

import (
	"google.golang.org/grpc"
)

func NewGrpcService(store MemoryStore) *grpc.Server {
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	metricsService := NewMetricsService(store)
	//traceService := NewTraceService()
	//logsService := NewLogsService()

	metricsService.Register(grpcServer)
	//traceService.Register(grpcServer)
	//logsService.Register(grpcServer)

	return grpcServer
}
