package main

import (
	"flag"
	"fmt"
	"github.com/gofiber/fiber/v3"
	"log"
	"net"
)

var (
	grpcPort = flag.Int("grpc_port", 4317, "The server's GRPC port")
	httpPort = flag.Int("http_port", 4318, "The server's HTTP port")
)

func main() {
	flag.Parse()

	app := fiber.New()

	metricsStore := NewMetricsStore()
	appService := AppService{metricsStore}

	app.Get("/", appService.Home)
	app.Get("/filter", appService.Filter)

	grpcService := NewGrpcService(metricsStore)

	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *grpcPort))
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		log.Printf("server listening at %v", lis.Addr())
		if err := grpcService.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *httpPort))
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		log.Printf("server listening at %v", lis.Addr())
		if err := grpcService.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)

		}
	}()

	log.Printf("server listening at %v", 8081)
	app.Listen("0.0.0.0:8081")

}
