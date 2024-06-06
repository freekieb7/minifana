package main

import (
	"github.com/gofiber/fiber/v3"
	v1 "go.opentelemetry.io/proto/otlp/metrics/v1"
)

type AppService interface {
	Home(c fiber.Ctx) error
	GetServiceNames(ctx fiber.Ctx) error
	GetServiceMetrics(ctx fiber.Ctx) error
	GetMetricNames(ctx fiber.Ctx) error
	GetTest(ctx fiber.Ctx) error
}

type appService struct {
	metricsStore MetricsStore
}

func NewAppService(metricsStore MetricsStore) AppService {
	return &appService{
		metricsStore: metricsStore,
	}
}

func (s *appService) Home(c fiber.Ctx) error {
	return c.Render("test", fiber.Map{
		"Options": nil,
		"Dates":   nil,
	})
}

func (s *appService) GetServiceNames(ctx fiber.Ctx) error {
	serviceNames := make([]string, 0)

	return ctx.JSON(serviceNames)
}

func (s *appService) GetServiceMetrics(ctx fiber.Ctx) error {
	metricNames := make([]string, 0)

	return ctx.JSON(metricNames)
}

func (s *appService) GetMetricNames(ctx fiber.Ctx) error {
	return ctx.Render("options", fiber.Map{
		"Options": s.metricsStore.Names(),
	})
}

func (s *appService) GetTest(ctx fiber.Ctx) error {
	var n string
	var v []*v1.NumberDataPoint

	for name, values := range s.metricsStore.Metrics() {
		n = name
		v = values
		break
	}

	return ctx.Render("more", fiber.Map{
		"Name":   n,
		"Values": v,
	})
}
