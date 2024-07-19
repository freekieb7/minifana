package main

import (
	"github.com/gofiber/fiber/v3"
	"log"
	"net/url"
	"strconv"
	"strings"
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
	startTime, _ := strconv.ParseUint(ctx.Query("s"), 10, 64)
	endTime, _ := strconv.ParseUint(ctx.Query("e"), 10, 64)

	log.Println(startTime)

	q := strings.Trim(ctx.Query("q"), "\"")
	queryValues, _ := url.ParseQuery(q)

	metrics, err := s.metricsStore.Filter(startTime, endTime, queryValues)

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
