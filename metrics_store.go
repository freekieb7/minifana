package main

import (
	_wal "github.com/tidwall/wal"
	_otlp_coll_v1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	_otlp_v1 "go.opentelemetry.io/proto/otlp/metrics/v1"
	"google.golang.org/protobuf/proto"
	"log"
)

type MetricsStore interface {
	AddMetrics(metrics []*_otlp_v1.ResourceMetrics)
	Names() []string
	Metrics() map[string][]*_otlp_v1.NumberDataPoint
}

type metricsStore struct {
	metrics []*_otlp_v1.ResourceMetrics
}

func NewMetricsStore() MetricsStore {
	wal, _ := _wal.Open("/minifana/metrics/wal", nil)

	metrics := make([]*_otlp_v1.ResourceMetrics, 0)

	var i uint64 = 1
	var req _otlp_coll_v1.ExportMetricsServiceRequest

	for {
		d, err := wal.Read(i)
		if err != nil {
			log.Println(err)
			break
		}

		i++

		if err := proto.Unmarshal(d, &req); err != nil {
			log.Println("shit2")

			continue
		}

		metrics = append(metrics, req.ResourceMetrics...)
	}

	return &metricsStore{
		metrics: metrics,
	}
}

func (ms *metricsStore) AddMetrics(metrics []*_otlp_v1.ResourceMetrics) {
	ms.metrics = append(ms.metrics, metrics...)
}

func (ms *metricsStore) Names() []string {
	names := make(map[string]bool)

	for _, metrics := range ms.metrics {
		for _, sm := range metrics.ScopeMetrics {
			for _, m := range sm.Metrics {
				names[m.Name] = true
			}
		}
	}

	keys := make([]string, len(names))

	i := 0
	for k := range names {
		keys[i] = k
		i++
	}

	return keys
}

func (ms *metricsStore) Metrics() map[string][]*_otlp_v1.NumberDataPoint {
	res := make(map[string][]*_otlp_v1.NumberDataPoint)

	for _, metrics := range ms.metrics {
		for _, sm := range metrics.ScopeMetrics {
			for _, m := range sm.Metrics {
				if s := m.GetSum(); s != nil {
					if _, ok := res[m.Name]; ok {
						res[m.Name] = append(res[m.Name], s.DataPoints...)
					} else {
						res[m.Name] = s.DataPoints
					}
				}
			}
		}
	}

	return res
}
