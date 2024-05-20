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
	port = flag.Int("port", 4317, "The server port")
)

func main() {
	flag.Parse()

	engine := html.New("./view", ".html")
	//engine.Reload(true)

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	ms := NewMemoryStore()

	appService := NewAppService(ms)

	app.Get("/", appService.Home)
	app.Get("/values", appService.GetMetricValues)
	//app.Get("/test", appService.Test)

	grpcService := NewGrpcService(ms)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("server listening at %v", lis.Addr())
	go func() {
		if err := grpcService.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	log.Printf("server listening at %v", 8081)
	app.Listen("0.0.0.0:8081")
	
}
