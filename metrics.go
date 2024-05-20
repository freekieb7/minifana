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
	store MemoryStore
	v1.UnimplementedMetricsServiceServer
}

func NewMetricsService(store MemoryStore) MetricsService {
	return &metricsService{
		store: store,
	}
}

func (s *metricsService) Register(grpcServer *grpc.Server) {
	v1.RegisterMetricsServiceServer(grpcServer, s)
}

func (s *metricsService) Export(_ context.Context, req *v1.ExportMetricsServiceRequest) (*v1.ExportMetricsServiceResponse, error) {
	// open a new log file
	log, err := wal.Open("/minifana/metrics/wal", nil)

	// write some entries
	li, _ := log.LastIndex()
	res, _ := proto.Marshal(req)

	err = log.Write(li+1, res)

	// close the log
	err = log.Close()

	if err != nil {
		return new(v1.ExportMetricsServiceResponse), err
	}

	serviceNames := make([]string, 0)

	for _, rm := range req.ResourceMetrics {
		for _, att := range rm.Resource.Attributes {
			if "service.name" == att.Key {
				serviceNames = append(serviceNames, att.Value.GetStringValue())
				//s.store.AddValue(att.Value.GetStringValue())
				//s.serviceMapping[att.Value.GetStringValue()] = append(s.serviceMapping[att.Value.GetStringValue()], id)
			}
		}

		for _, sm := range rm.ScopeMetrics {
			for _, m := range sm.Metrics {
				s.store.AddValue(m.Name)
			}
		}
	}

	return new(v1.ExportMetricsServiceResponse), nil
}
