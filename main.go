package main

import (
	"fmt"
	"github.com/acontrolfreak/minifana/metrics"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"os"
)

const (
	telemetryGrpcPort = 4317
	telemetryHttpPort = 4318
	apiPort           = 8081
)

func main() {
	grpcServer := grpc.NewServer()

	metricsStorageFile, err := os.CreateTemp("", "metrics_storage*.bin")
	//metricsStorageFile, err := os.Create("/tmp/metrics_storage3797485148.bin")

	if err != nil {
		panic(err)
	}

	log.Println(metricsStorageFile.Name())

	defer os.Remove(metricsStorageFile.Name())

	metricsController := metrics.Controller{
		StorageFile: metricsStorageFile,
	}
	metricsController.Register(grpcServer)

	//traceService := NewTraceService()
	//logsService := NewLogsService()

	//traceService.Register(grpcServer)
	//logsService.Register(grpcServer)

	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", telemetryGrpcPort))
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		log.Printf("server listening at %v", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", telemetryHttpPort))
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		log.Printf("server listening at %v", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)

		}
	}()

	log.Printf("server listening at %v", apiPort)

	http.HandleFunc("/hello", func(writer http.ResponseWriter, request *http.Request) {
		metricsController.Read()
	})
	http.ListenAndServe(fmt.Sprintf(":%d", apiPort), nil)
}
