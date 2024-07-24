package metrics

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	collectorProtoMetrics "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	protoMetrics "go.opentelemetry.io/proto/otlp/metrics/v1"
	"google.golang.org/grpc"
	"io"
	"log"
	"math"
	"os"
	"sync"

	"google.golang.org/protobuf/proto"
)

const (
	MetricData_GAUGE                 = 0x0
	MetricData_SUM                   = 0x1
	MetricData_HISTOGRAM             = 0x2
	MetricData_EXPONENTIAL_HISTOGRAM = 0x3
	MetricData_SUMMARY               = 0x4

	AggregationTemporality_AGGREGATION_TEMPORALITY_UNSPECIFIED = 0x0
	AggregationTemporality_AGGREGATION_TEMPORALITY_DELTA       = 0x1
	AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE  = 0x2

	Monotonic_IS_NOT_MONOTONIC = 0x0
	Monotonic_IS_MONOTONIC     = 0x1
)

type Controller struct {
	collectorProtoMetrics.UnimplementedMetricsServiceServer
	StorageFile *os.File
	mutex       sync.Mutex
}

func (c *Controller) Register(grpcServer *grpc.Server) {
	collectorProtoMetrics.RegisterMetricsServiceServer(grpcServer, c)
}

func (c *Controller) Read() {
	log.Println("Reading metrics")
	metaData := make([]byte, 1)
	dataPointLengthInBytes := make([]byte, 2)

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, err := c.StorageFile.Seek(0, 0); err != nil {
		log.Println(err)
		return
	} // Reset file pointer to start of file

	for {
		if _, err := c.StorageFile.Read(metaData); err != nil {
			if err != io.EOF {
				log.Println(err)
			}

			return
		}

		if _, err := c.StorageFile.Read(dataPointLengthInBytes); err != nil {
			log.Println(err)
			return
		}

		dataPointLength := binary.LittleEndian.Uint16(dataPointLengthInBytes)
		dataPointInBytes := make([]byte, dataPointLength)

		if _, err := c.StorageFile.Read(dataPointInBytes); err != nil {
			log.Println(err)
			return
		}

		if metaData[0]&MetricData_SUMMARY == MetricData_SUMMARY {
			var dataPoint protoMetrics.SummaryDataPoint
			proto.Unmarshal(dataPointInBytes, &dataPoint)
			log.Println(dataPoint.TimeUnixNano)
		} else if metaData[0]&MetricData_EXPONENTIAL_HISTOGRAM == MetricData_EXPONENTIAL_HISTOGRAM {
			var dataPoint protoMetrics.ExponentialHistogramDataPoint
			proto.Unmarshal(dataPointInBytes, &dataPoint)
			log.Println(dataPoint.TimeUnixNano)
		} else if metaData[0]&MetricData_HISTOGRAM == MetricData_HISTOGRAM {
			var dataPoint protoMetrics.HistogramDataPoint
			proto.Unmarshal(dataPointInBytes, &dataPoint)
			log.Println(dataPoint.TimeUnixNano)
		} else if metaData[0]&MetricData_SUM == MetricData_SUM {
			var dataPoint protoMetrics.NumberDataPoint
			proto.Unmarshal(dataPointInBytes, &dataPoint)
			log.Println(dataPoint.TimeUnixNano)
		} else if metaData[0]&MetricData_GAUGE == MetricData_GAUGE {
			var dataPoint protoMetrics.NumberDataPoint
			proto.Unmarshal(dataPointInBytes, &dataPoint)
			log.Println(dataPoint.TimeUnixNano)
		} else {
			log.Printf("Unknown metric type in metadata: %b\n", metaData[0])
		}
	}
}

func (c *Controller) Export(_ context.Context, req *collectorProtoMetrics.ExportMetricsServiceRequest) (*collectorProtoMetrics.ExportMetricsServiceResponse, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, resourceMetrics := range req.ResourceMetrics {
		for _, scopeMetric := range resourceMetrics.ScopeMetrics {
			for _, metric := range scopeMetric.Metrics {
				if val, ok := metric.Data.(*protoMetrics.Metric_Gauge); ok {
					metaData := make([]byte, 1)

					/* Metric type (3 bits) */
					metaData[0] |= MetricData_GAUGE

					for _, dataPoint := range val.Gauge.DataPoints {
						if err := c.writeToFile(dataPoint, metaData); err != nil {
							log.Println(err)
						}
					}

					break
				}

				if val, ok := metric.Data.(*protoMetrics.Metric_Sum); ok {
					metaData := make([]byte, 1)

					/* Is monotonic (1 bit) */
					if val.Sum.IsMonotonic {
						metaData[0] |= Monotonic_IS_MONOTONIC
					}

					/* Aggregation type (2 bits) */
					metaData[0] = metaData[0] << 2 // Make room
					if val.Sum.AggregationTemporality == protoMetrics.AggregationTemporality_AGGREGATION_TEMPORALITY_UNSPECIFIED {
						metaData[0] |= AggregationTemporality_AGGREGATION_TEMPORALITY_UNSPECIFIED
					} else if val.Sum.AggregationTemporality == protoMetrics.AggregationTemporality_AGGREGATION_TEMPORALITY_DELTA {
						metaData[0] |= AggregationTemporality_AGGREGATION_TEMPORALITY_DELTA
					} else if val.Sum.AggregationTemporality == protoMetrics.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE {
						metaData[0] |= AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE
					}

					/* Metric type (3 bits) */
					metaData[0] = metaData[0] << 3 // Make room
					metaData[0] |= MetricData_SUM

					for _, dataPoint := range val.Sum.DataPoints {
						if err := c.writeToFile(dataPoint, metaData); err != nil {
							log.Println(err)
						}
					}

					break
				}

				if val, ok := metric.Data.(*protoMetrics.Metric_Histogram); ok {
					metaData := make([]byte, 1)

					/* Aggregation type (2 bits) */
					if val.Histogram.AggregationTemporality == protoMetrics.AggregationTemporality_AGGREGATION_TEMPORALITY_UNSPECIFIED {
						metaData[0] |= AggregationTemporality_AGGREGATION_TEMPORALITY_UNSPECIFIED
					} else if val.Histogram.AggregationTemporality == protoMetrics.AggregationTemporality_AGGREGATION_TEMPORALITY_DELTA {
						metaData[0] |= AggregationTemporality_AGGREGATION_TEMPORALITY_DELTA
					} else if val.Histogram.AggregationTemporality == protoMetrics.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE {
						metaData[0] |= AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE
					}

					/* Metric type (3 bits) */
					metaData[0] = metaData[0] << 3 // Make room
					metaData[0] |= MetricData_HISTOGRAM
					for _, dataPoint := range val.Histogram.DataPoints {
						if err := c.writeToFile(dataPoint, metaData); err != nil {
							log.Println(err)
						}
					}

					break
				}

				if val, ok := metric.Data.(*protoMetrics.Metric_ExponentialHistogram); ok {
					metaData := make([]byte, 1)

					/* Aggregation type (2 bits) */
					if val.ExponentialHistogram.AggregationTemporality == protoMetrics.AggregationTemporality_AGGREGATION_TEMPORALITY_UNSPECIFIED {
						metaData[0] |= AggregationTemporality_AGGREGATION_TEMPORALITY_UNSPECIFIED
					} else if val.ExponentialHistogram.AggregationTemporality == protoMetrics.AggregationTemporality_AGGREGATION_TEMPORALITY_DELTA {
						metaData[0] |= AggregationTemporality_AGGREGATION_TEMPORALITY_DELTA
					} else if val.ExponentialHistogram.AggregationTemporality == protoMetrics.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE {
						metaData[0] |= AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE
					}

					/* Metric type (3 bits) */
					metaData[0] = metaData[0] << 3 // Make room
					metaData[0] |= MetricData_EXPONENTIAL_HISTOGRAM

					for _, dataPoint := range val.ExponentialHistogram.DataPoints {
						if err := c.writeToFile(dataPoint, metaData); err != nil {
							log.Println(err)
						}
					}

					break
				}

				if val, ok := metric.Data.(*protoMetrics.Metric_Summary); ok {
					metaData := make([]byte, 1)

					/* Metric type (3 bits) */
					metaData[0] |= MetricData_SUMMARY

					for _, dataPoint := range val.Summary.DataPoints {
						if err := c.writeToFile(dataPoint, metaData); err != nil {
							log.Println(err)
						}
					}

					break
				}
			}
		}
	}

	return new(collectorProtoMetrics.ExportMetricsServiceResponse), nil
}

func (c *Controller) writeToFile(dataPoint proto.Message, metaData []byte) error {
	s, _ := c.StorageFile.Stat()

	if s.Size() != 0 {
		return nil
	}

	if dataPoint == nil {
		return errors.New("dataPoint is nil")
	}

	dataPointInBytes, err := proto.Marshal(dataPoint)
	if err != nil {
		log.Println("datapoint marshal went wrong", err)
	}

	size := len(dataPointInBytes)
	if size > math.MaxUint16 {
		return fmt.Errorf("message is bigger then max %d", math.MaxUint16)
	}

	dataPointLengthInBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(dataPointLengthInBytes, uint16(size))

	c.StorageFile.Write(metaData)
	c.StorageFile.Write(dataPointLengthInBytes)
	c.StorageFile.Write(dataPointInBytes)

	return err
}
