package main

import (
	"flag"
	"fmt"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/template/html/v2"
	"log"
	"net"
)

var (
	grpcPort = flag.Int("grpc_port", 4317, "The server's GRPC port")
	httpPort = flag.Int("http_port", 4318, "The server's HTTP port")
)

func main() {
	flag.Parse()

	engine := html.New("./view", ".html")
	//engine.Reload(true)

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	ms := NewMetricsStore()

	appService := NewAppService(ms)

	app.Get("/", appService.Home)
	app.Get("/metrics/names", appService.GetMetricNames)
	app.Get("/metrics/test", appService.GetTest)

	grpcService := NewGrpcService(ms)

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
