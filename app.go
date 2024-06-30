package main

import (
	"github.com/gofiber/fiber/v3"
)

type AppService struct {
	metricsStore *MetricsStore
}

func (s *AppService) Home(ctx fiber.Ctx) error {
	m, err := s.metricsStore.Metrics()

	if err != nil {
		return err
	}

	return ctx.JSON(m)
}

func (s *AppService) Filter(ctx fiber.Ctx) error {
	name := ctx.Query("name")
	value := ctx.Query("value")

	metrics, err := s.metricsStore.Filter(name, value)

	if err != nil {
		return err
	}

	return ctx.JSON(metrics)
}

func (s *AppService) GetServiceNames(ctx fiber.Ctx) error {
	serviceNames := make([]string, 0)

	return ctx.JSON(serviceNames)
}

func (s *AppService) GetServiceMetrics(ctx fiber.Ctx) error {
	metricNames := make([]string, 0)

	return ctx.JSON(metricNames)
}
