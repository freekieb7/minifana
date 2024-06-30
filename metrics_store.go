package main

import (
	"github.com/tidwall/wal"
	collector_v1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	v12 "go.opentelemetry.io/proto/otlp/common/v1"
	v1 "go.opentelemetry.io/proto/otlp/metrics/v1"
	"google.golang.org/protobuf/proto"
	"log"
)

const (
	walFileLocation = "/minifana/metrics/wal"
)

type MetricsStore struct {
	metrics []*v1.ResourceMetrics
	walFile *wal.Log
}

func NewMetricsStore() *MetricsStore {
	walFile, _ := wal.Open(walFileLocation, nil)

	metrics := make([]*v1.ResourceMetrics, 0)

	var i uint64 = 1
	var req collector_v1.ExportMetricsServiceRequest

	for {
		d, err := walFile.Read(i)
		if err != nil {
			log.Println(err)
			break
		}

		i++

		if err := proto.Unmarshal(d, &req); err != nil {
			continue
		}

		metrics = append(metrics, req.ResourceMetrics...)
	}

	//walFile.Close()

	return &MetricsStore{
		metrics: metrics,
		walFile: walFile,
	}
}

func (store *MetricsStore) AddToWal(req *collector_v1.ExportMetricsServiceRequest) error {
	li, err := store.walFile.LastIndex()

	if err != nil {
		return err
	}

	res, err := proto.Marshal(req)

	if err != nil {
		return err
	}

	err = store.walFile.Write(li+1, res)

	if err != nil {
		return err
	}

	// Add metrics to memory
	store.metrics = append(store.metrics, req.ResourceMetrics...)
	return nil
}

func (store *MetricsStore) Filter(name string, value string) ([]*Metric, error) {
	filtered := make([]*Metric, 0)
	metrics, err := store.Metrics()

	if err != nil {
		return filtered, err
	}

	for _, metric := range metrics {
		if val, ok := metric.Attributes[name]; ok {
			log.Println(val)
			if val == value {

				filtered = append(filtered, metric)
			}
		}
	}

	return filtered, nil
}

type Metric struct {
	Name       string         `json:"n"`
	Attributes map[string]any `json:"a"`
	Values     []MetricValue  `json:"v"`
}

type MetricValue struct {
	Value      any            `json:"v"`
	Timestamp  uint64         `json:"t"`
	Attributes map[string]any `json:"a,omitempty"`
}

func (store *MetricsStore) Metrics() ([]*Metric, error) {
	flatMetrics := make([]*Metric, 0, len(store.metrics))

	for _, metrics := range store.metrics {
		var flatMetric Metric
		if res := metrics.Resource; res != nil {
			if res.Attributes != nil {
				flatMetric.Attributes = attributesAsList(res.Attributes)
			}
		}

		for _, sm := range metrics.ScopeMetrics {
			for _, m := range sm.Metrics {
				flatMetric.Name = m.Name

				addValueToFlatMetric(&flatMetric, m)

				flatMetrics = append(flatMetrics, &flatMetric)
			}
		}
	}

	return flatMetrics, nil
}

func addValueToFlatMetric(flatMetric *Metric, m *v1.Metric) {
	if metric := m.GetSum(); metric != nil {
		flatMetric.Name += "_sum"

		for _, dp := range metric.DataPoints {
			if x, ok := dp.GetValue().(*v1.NumberDataPoint_AsDouble); ok {
				flatMetric.Values = append(flatMetric.Values, MetricValue{
					Value:      x.AsDouble,
					Timestamp:  dp.TimeUnixNano,
					Attributes: attributesAsList(dp.Attributes),
				})
				return
			}

			if x, ok := dp.GetValue().(*v1.NumberDataPoint_AsInt); ok {
				flatMetric.Values = append(flatMetric.Values, MetricValue{
					Value:      x.AsInt,
					Timestamp:  dp.TimeUnixNano,
					Attributes: attributesAsList(dp.Attributes),
				})
				return
			}

			log.Fatal("no case match for sum metric value")
		}
	}

	if metric := m.GetSummary(); metric != nil {
		flatMetric.Name += "_summary_sum"
		for _, dp := range metric.DataPoints {
			flatMetric.Values = append(flatMetric.Values, MetricValue{
				Value:      dp.Sum,
				Timestamp:  dp.TimeUnixNano,
				Attributes: attributesAsList(dp.Attributes),
			})
		}
		return
	}

	if metric := m.GetHistogram(); metric != nil {
		flatMetric.Name += "_histogram_sum"
		for _, dp := range metric.DataPoints {
			flatMetric.Values = append(flatMetric.Values, MetricValue{
				Value:      dp.Sum,
				Timestamp:  dp.TimeUnixNano,
				Attributes: attributesAsList(dp.Attributes),
			})
		}

		return
	}

	if metric := m.GetGauge(); metric != nil {
		flatMetric.Name += "_gauge"
		for _, dp := range metric.DataPoints {
			if x, ok := dp.GetValue().(*v1.NumberDataPoint_AsDouble); ok {
				flatMetric.Values = append(flatMetric.Values, MetricValue{
					Value:      x.AsDouble,
					Timestamp:  dp.TimeUnixNano,
					Attributes: attributesAsList(dp.Attributes),
				})
				return
			}

			if x, ok := dp.GetValue().(*v1.NumberDataPoint_AsInt); ok {
				flatMetric.Values = append(flatMetric.Values, MetricValue{
					Value:      x.AsInt,
					Timestamp:  dp.TimeUnixNano,
					Attributes: attributesAsList(dp.Attributes),
				})
				return
			}

			log.Fatal("no case match for gauge metric value")
		}
	}

	if metric := m.GetExponentialHistogram(); metric != nil {
		flatMetric.Name += "_exponential_histogram_sum"
		for _, dp := range metric.DataPoints {
			flatMetric.Values = append(flatMetric.Values, MetricValue{
				Value:      dp.Sum,
				Timestamp:  dp.TimeUnixNano,
				Attributes: attributesAsList(dp.Attributes),
			})
		}
		return
	}
}

func attributesAsList(keyValuePairs []*v12.KeyValue) map[string]any {
	attributes := make(map[string]any)

	for _, item := range keyValuePairs {
		if item.Value == nil {
			continue
		}

		if x, ok := item.Value.Value.(*v12.AnyValue_StringValue); ok {
			attributes[item.Key] = x.StringValue
			continue
		}

		if x, ok := item.Value.Value.(*v12.AnyValue_BoolValue); ok {
			attributes[item.Key] = x.BoolValue
			continue
		}

		if x, ok := item.Value.Value.(*v12.AnyValue_IntValue); ok {
			attributes[item.Key] = x.IntValue
			continue
		}

		if x, ok := item.Value.Value.(*v12.AnyValue_DoubleValue); ok {
			attributes[item.Key] = x.DoubleValue
			continue
		}

		if x, ok := item.Value.Value.(*v12.AnyValue_ArrayValue); ok {
			attributes[item.Key] = x.ArrayValue
			continue
		}

		if x, ok := item.Value.Value.(*v12.AnyValue_KvlistValue); ok {
			attributes[item.Key] = x.KvlistValue
			continue
		}

		if x, ok := item.Value.Value.(*v12.AnyValue_BytesValue); ok {
			attributes[item.Key] = x.BytesValue
			continue
		}
	}

	return attributes
}
