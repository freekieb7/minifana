package main

import (
	"context"
	"github.com/tidwall/wal"
	_ "go.beyondstorage.io/services/fs/v4"
	v1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type MetricsService interface {
	v1.MetricsServiceServer
	Register(grpcServer *grpc.Server)
}

type metricsService struct {
	store MetricsStore
	v1.UnimplementedMetricsServiceServer
}

func NewMetricsService(store MetricsStore) MetricsService {
	return &metricsService{
		store: store,
	}
}

func (s *metricsService) Register(grpcServer *grpc.Server) {
	v1.RegisterMetricsServiceServer(grpcServer, s)
}

func (s *metricsService) Export(_ context.Context, req *v1.ExportMetricsServiceRequest) (*v1.ExportMetricsServiceResponse, error) {
	if err := writeToWal(req); err != nil {
		return new(v1.ExportMetricsServiceResponse), err
	}

	s.store.AddMetrics(req.ResourceMetrics)

	return new(v1.ExportMetricsServiceResponse), nil
}

func writeToWal(req *v1.ExportMetricsServiceRequest) error {
	// open a new log file
	log, err := wal.Open("/minifana/metrics/wal", nil)

	if err != nil {
		return err
	}

	li, err := log.LastIndex()

	if err != nil {
		return err
	}

	res, err := proto.Marshal(req)

	if err != nil {
		return err
	}

	err = log.Write(li+1, res)

	if err != nil {
		return err
	}

	return log.Close()
}
