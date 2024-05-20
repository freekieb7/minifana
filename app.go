package main

import (
	"github.com/gofiber/fiber/v3"
)

type AppService interface {
	Home(c fiber.Ctx) error
	GetServiceNames(ctx fiber.Ctx) error
	GetServiceMetrics(ctx fiber.Ctx) error
	GetMetricValues(ctx fiber.Ctx) error
}

type appService struct {
	store MemoryStore
}

func NewAppService(store MemoryStore) AppService {
	return &appService{
		store: store,
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

func (s *appService) GetMetricValues(ctx fiber.Ctx) error {
	return ctx.JSON(s.store.GetValues())
}
