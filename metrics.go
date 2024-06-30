package main

import (
	"context"
	_ "go.beyondstorage.io/services/fs/v4"
	v1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"google.golang.org/grpc"
)

type MetricsService interface {
	v1.MetricsServiceServer
	Register(grpcServer *grpc.Server)
}

type metricsService struct {
	store *MetricsStore
	v1.UnimplementedMetricsServiceServer
}

func NewMetricsService(store *MetricsStore) MetricsService {
	return &metricsService{
		store: store,
	}
}

func (s *metricsService) Register(grpcServer *grpc.Server) {
	v1.RegisterMetricsServiceServer(grpcServer, s)
}

func (s *metricsService) Export(_ context.Context, req *v1.ExportMetricsServiceRequest) (*v1.ExportMetricsServiceResponse, error) {
	if err := s.store.AddToWal(req); err != nil {
		return new(v1.ExportMetricsServiceResponse), err
	}

	return new(v1.ExportMetricsServiceResponse), nil
}
